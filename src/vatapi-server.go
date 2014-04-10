package main

import (
	"fmt"
	"net/http"
	"net/url"
	"encoding/csv"
	"os"
	"io"
	"io/ioutil"
	"log"
	"flag"
	"time"
	"strconv"
	"crypto/rand"
	"math/big"
	_ "github.com/lib/pq"
	"database/sql"
	"github.com/go-martini/martini"
	"github.com/satori/go.uuid"
	"github.com/sirsean/go-mailgun/mailgun"
	// "github.com/martini-contrib/auth"
)

var taxFlag = flag.String("tax", "", "CSV with taxes")
var dbFlag = flag.String("db", "", "Postgres configuration")
func init() {
	// example with short version for long flag
	flag.StringVar(taxFlag, "t", "", "CSV with taxes")
	flag.StringVar(dbFlag, "d", "", "Postgres configuration")
}

type Data struct {
	Tax map[string]string
	db *sql.DB
	mailgunClient *mailgun.Client

}

func (d *Data) calcHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.RawQuery
	q, _ := url.ParseQuery(query)
	country := q["country"][0]
	callbackparam := q["callback"]
	apikeyparam := q["apikey"]
	amount, _ := strconv.Atoi(q["amount"][0])
	if callbackparam == nil {
	fmt.Fprintf(w, "{")
	} else {
		callback := callbackparam[0]
		fmt.Fprintf(w, "%s({", callback)
	}
	success := false
	taxrate := d.Tax[country]
	if apikeyparam == nil {

		fmt.Fprintf(w, "\"error\": \"%s\", ", "Missing apikey")
	} else if taxrate == "" {
		error := "Unknown country"
		fmt.Fprintf(w, "\"error\": \"%s\", ", error)
	} else {
		//Check API key
		apikey := apikeyparam[0]
		rows := d.db.QueryRow("SELECT id FROM client WHERE apikey=$1",
			apikey)
		var clientid string
		if err := rows.Scan(&clientid); err != nil {
			log.Println(err)
		}
		if clientid == "" {
			fmt.Fprintf(w, "\"error\": %s, ", "Wrong API key")
		} else {
			amountWithTax, tax := d.calcAmount(amount, country)
			success = true
			fmt.Fprintf(w, "\"taxrate\": %s, ", taxrate)
			fmt.Fprintf(w, "\"amount\": %d, ", amount)
			fmt.Fprintf(w, "\"tax\": %d, ", tax)
			fmt.Fprintf(w, "\"amountwithtax\": %d, ", amountWithTax)
			fmt.Fprintf(w, "\"country\": \"%s\", ", country)
		}
	}


	fmt.Fprintf(w, "\"success\": %t ", success)

	if callbackparam == nil {
		fmt.Fprintf(w, "}\n")
	} else {
		fmt.Fprintf(w, "});\n")
	}

}

func (d *Data) calcAmount(amnt int, country string) (int, int) {
	taxrate, _ := strconv.Atoi(d.Tax[country])

	tax := amnt * taxrate / 100
	amntWithTax := amnt + tax
	return amntWithTax, tax
}

func (d *Data) saleHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.RawQuery
	q, _ := url.ParseQuery(query)
	country := q["country"][0]
	callbackparam := q["callback"]
	apikeyparam := q["apikey"]
	//log.Println(query)
	amount, _ := strconv.Atoi(q["amount"][0])
	if callbackparam == nil {
		fmt.Fprintf(w, "{")
	} else {
		callback := callbackparam[0]
		fmt.Fprintf(w, "%s({", callback)
	}
	success := false
	taxrate := d.Tax[country]
	var amountWithTax int
	var tax int
	if apikeyparam == nil {

		fmt.Fprintf(w, "\"error\": \"%s\",", "Missing apikey")
	} else if taxrate == "" {

		error := "Unknown country"
		fmt.Fprintf(w, "\"error\": \"%s\",", error)
	} else {
		//Check API key
		apikey := apikeyparam[0]
		rows := d.db.QueryRow("SELECT id FROM client WHERE apikey=$1",
			apikey)
		var clientid string
		if err := rows.Scan(&clientid); err != nil {
			log.Println(err)
		}
		if clientid == "" {
			fmt.Fprintf(w, "\"error\": %s,", "Wrong API key")
		} else {
			timestampstring := q["timestamp"][0]
			timestampint, _ := strconv.ParseInt(timestampstring, 10, 64)
			timestamp := time.Unix(timestampint, 0)
			amountWithTax, tax = d.calcAmount(amount, country)
			success = true
			fmt.Fprintf(w, "\"taxrate\": %s,", taxrate)
			fmt.Fprintf(w, "\"amount\": %d,", amount)
			fmt.Fprintf(w, "\"tax\": %d,", tax)
			fmt.Fprintf(w, "\"amountwithtax\": %d,", amountWithTax)
			timenow := time.Now()
			_, err := d.db.Exec("INSERT INTO sale (amount, tax, amountwithtax, taxrate, countrycode, timestamp, createdat, client_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)",
				amount, tax, amountWithTax, taxrate, country, timestamp, timenow, clientid)

			if err != nil {
				fmt.Fprintf(w, "\"error\": \"%s\",", err)
			}
			fmt.Fprintf(w, "\"country\": \"%s\",", country)
			success = true
		}
	}


	fmt.Fprintf(w, "\"success\": %t", success)


	if callbackparam == nil {
		fmt.Fprintf(w, "}\n")
	} else {
		fmt.Fprintf(w, "});\n")
	}



}

func readParam(param []string) string {
	if (param != nil) {
		return param[0]
	}
	return ""
}

func (d *Data) signup(w http.ResponseWriter, r *http.Request) {
	query := r.URL.RawQuery
	q, _ := url.ParseQuery(query)
	callbackparam := q["callback"]
	nameparam := q["name"]
	emailparam := q["email"]
	companynameparam := q["companyname"]
	addressparam := q["address"]
	cityparam := q["city"]
	zipparam := q["zip"]
	stateparam := q["state"]
	countryparam := q["country"]
	passwordparam := q["password"]
	planparam := q["plan"]
	success := false
	if callbackparam == nil && nameparam == nil && emailparam == nil && companynameparam == nil && passwordparam == nil && planparam == nil {
		fmt.Fprintf(w, "\"error\": \"%s\",\n", "Parameters not sufficient")
	} else {
		callback, name, email, companyname, password, plan := callbackparam[0], nameparam[0], emailparam[0], companynameparam[0], passwordparam[0], planparam[0]
		address, city, zip, state, country := readParam(addressparam), readParam(cityparam), readParam(zipparam), readParam(stateparam), readParam(countryparam)
		var querylimit int
		taxreports, trial, confirmed := false, true, false
		switch plan {
		case "free":
			querylimit = 1000
		case "basic":
			querylimit = 200000
		case "professional":
			querylimit = 500000   //Unlimited
			taxreports = true
		case "premium":
			querylimit = 10000000   //Unlimited
			taxreports = true
		}
		fmt.Fprintf(w, "%s({\n", callback)
		//Generate confirmation email
		confirmkey := randString(64)
		confirmkeyvalid := time.Now().AddDate(0, 0, 7)

		message := mailgun.Message{
			FromName:    "VAT API",
			FromAddress: "support@vatapi.co",
			ToAddress:   email,
			BCCAddressList: []string{"mislav@infolution.biz"},
			Subject:     "VAT API signup confirmation",
			Body:        `You registered for VAT API.
			To start using it, go to our website
			http://www.vatapi.co
			and get familiar with code examples.

			We are always making our product better, so if you run into some problems with
			implementation or any other issue that's bothering you feel free to contact us at
			support@vatapi.co

			Thank you,
			Mislav Kasner
			VAT API founder`,
		}

		body, err := d.mailgunClient.Send(message)
		if err != nil {
			fmt.Fprintf(w, "\"error\": \"%s\",\n", "Error sending confirmation e-mail")
		} else {
			fmt.Fprintf(w, "\"confirmationmail\": \"%s\",\n", body)
		}


		trialend := time.Now().AddDate(0, 0, 30)
		u1 := uuid.NewV4()
		apikey := randString(32)




		_ = time.Now()

		_, err = d.db.Exec(`INSERT INTO client (id, name, address, zip, city, state, country, apikey, plan, querylimit, taxreports, trial, trialend, confirmed, confirmkey, confirmkeyvalid)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
			u1.String(), companyname, address, zip, city, state, country, apikey, plan, querylimit, taxreports, trial, trialend, confirmed, confirmkey, confirmkeyvalid)

		if err != nil {
			fmt.Fprintf(w, "\"error\": \"%s\",\n", err)
		}

		active := true


		_, err = d.db.Exec("INSERT INTO loginuser (name, email, password, active, client_id) VALUES ($1,$2,$3,$4,$5)",
			name, email, password, active, u1.String())

		if err != nil {
			fmt.Fprintf(w, "\"error\": \"%s\",\n", err)
		} else {
			success = true
		}
	}

	fmt.Fprintf(w, "\"success\": %t,\n", success)
	fmt.Fprintf(w, "});\n")



}

func randString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	symbols := big.NewInt(int64(len(alphanum)))
	states := big.NewInt(0)
	states.Exp(symbols, big.NewInt(int64(n)), nil)
	r, err := rand.Int(rand.Reader, states)
	if err != nil {
		panic(err)
	}
	var bytes = make([]byte, n)
	r2 := big.NewInt(0)
	symbol := big.NewInt(0)
	for i := range bytes {
		r2.DivMod(r, symbols, symbol)
		r, r2 = r2, r
		bytes[i] = alphanum[symbol.Int64()]
	}
	return string(bytes)
}

func (d *Data) generateAPIKey(w http.ResponseWriter, r *http.Request) {
	query := r.URL.RawQuery
	q, _ := url.ParseQuery(query)
	callback := q["callback"][0]
	clientid := q["clientid"][0]
	session := q["session"][0]
	success := false
	fmt.Fprintf(w, "%s({\n", callback)
	if session == "" {
		fmt.Fprintf(w, "\"error\": \"%s\",\n", "Session not provided")
	} else {
		apikey := randString(32)
		_, err := d.db.Exec("UPDATE client SET apikey=$1 WHERE id=$2",
			apikey, clientid)

		if err != nil {
			fmt.Fprintf(w, "\"error\": \"%s\",\n", err)
		} else {
			success = true
		}
		fmt.Fprintf(w, "\"apikey\": \"%s\",\n", apikey)
	}






	//	if err != nil {
	//		fmt.Fprintf(w, "\"error\": \"%s\",\n", err)
	//	} else {
	//		success = true
	//	}

	fmt.Fprintf(w, "\"success\": %t,\n", success)
	fmt.Fprintf(w, "});\n")
}

func (d *Data) login(w http.ResponseWriter, r *http.Request) {
	query := r.URL.RawQuery
	q, _ := url.ParseQuery(query)
	callback := q["callback"][0]
	email := q["email"][0]
	password := q["password"][0]
	session := randString(32)

	rows := d.db.QueryRow("SELECT client_id FROM loginuser WHERE email=$1 and password=$2",
		email, password)
	var clientid string
	if err := rows.Scan(&clientid); err != nil {
		log.Panic(err)
	}

	success := true
	fmt.Fprintf(w, "%s({\n", callback)
	fmt.Fprintf(w, "\"clientid\": \"%s\",\n", clientid)
	fmt.Fprintf(w, "\"session\": \"%s\",\n", session)


	//	if err != nil {
	//		fmt.Fprintf(w, "\"error\": \"%s\",\n", err)
	//	} else {
	//		success = true
	//	}

	fmt.Fprintf(w, "\"success\": %t,\n", success)
	fmt.Fprintf(w, "});\n")
}

func (d *Data) accountInfo(w http.ResponseWriter, r *http.Request) {
	query := r.URL.RawQuery
	q, _ := url.ParseQuery(query)
	callback := q["callback"][0]
	clientid := q["clientid"][0]
	session := q["session"][0]
	success := false
	fmt.Fprintf(w, "%s({\n", callback)
	if session == "" {
		fmt.Fprintf(w, "\"error\": \"%s\",\n", "Session not provided")
	} else {

		rows := d.db.QueryRow(`SELECT client.name, client.address, client.zip, client.city, client.state, client.country, client.apikey, loginuser.email a FROM client
								LEFT JOIN loginuser ON loginuser.client_id=client.id
								WHERE client.id=$1`,
			clientid)
		var name, address, zip, country, city, state, apikey, email string
		if err := rows.Scan(&name, &address, &zip, &city, &state, &country, &apikey, &email); err != nil {
			log.Panic(err)
		}

		success = true

		fmt.Fprintf(w, "\"name\": \"%s\",\n", name)
		fmt.Fprintf(w, "\"address\": \"%s\",\n", address)
		fmt.Fprintf(w, "\"zip\": \"%s\",\n", zip)
		fmt.Fprintf(w, "\"city\": \"%s\",\n", city)
		fmt.Fprintf(w, "\"state\": \"%s\",\n", state)
		fmt.Fprintf(w, "\"country\": \"%s\",\n", country)
		fmt.Fprintf(w, "\"apikey\": \"%s\",\n", apikey)
		fmt.Fprintf(w, "\"email\": \"%s\",\n", email)
	}

	//	if err != nil {
	//		fmt.Fprintf(w, "\"error\": \"%s\",\n", err)
	//	} else {
	//		success = true
	//	}

	fmt.Fprintf(w, "\"success\": %t,\n", success)
	fmt.Fprintf(w, "});\n")



}

func (d *Data) testdbHandler() string {
	//age := 21
	//rows, err := d.db.Query("SELECT countrycode FROM sale WHERE amount > $1", age)
	var name string
	rows := d.db.QueryRow("SELECT 'mislav' as name")
	if err := rows.Scan(&name); err != nil {
		log.Panic(err)
	}

	return "Number of rows " + name
}

func readTaxes() map[string]string {

	if *taxFlag == "" {
		log.Fatal("CSV with taxes is missing")
	}
	csvFile, err := os.Open(*taxFlag)
	defer csvFile.Close()
	if err != nil {
		panic(err)
	}

	csvReader := csv.NewReader(csvFile)
	m := make(map[string]string)
	for {
		fields, err := csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		//fmt.Printf("%s\n",fields[2])

		m[fields[0]] = fields[2]
	}
	return m
}

func main() {
	flag.Parse()
	m := readTaxes()  //Map with taxes

	//Mailgun
	mailgunClient := mailgun.NewClient("key-1tbrmvpihgzt2hzac-pguge25zx8m0w4", "vatapi.co")

	//Database
	if *dbFlag == "" {
		log.Fatal("Postgres configuration is missing")
	}
	buff, err := ioutil.ReadFile(*dbFlag)
	if err != nil {
		log.Fatal(err)
	}
	connstring := string(buff)
	//fmt.Printf("%s\n", connstring)
	db, err := sql.Open("postgres", connstring)

	if err != nil {
		log.Fatal(err)
	}
	//defer db.Close()
	d := &Data{Tax: m, db: db, mailgunClient: mailgunClient}
	//fmt.Printf("%s\n", m["HR"])
	r := martini.Classic()

	//r.Use(auth.Basic("username", "secretpassword"))
	r.Get("/calc/", d.calcHandler)
	r.Get("/sale/", d.saleHandler)
	r.Get("/testdb/", d.testdbHandler)
	r.Get("/signup/", d.signup)
	r.Get("/generateapi/", d.generateAPIKey)
	r.Get("/account/", d.accountInfo)
	r.Get("/login/", d.login)
	r.Run()
	//http.HandleFunc("/", d.handler)
	//log.Fatal(http.ListenAndServe(":8080", nil))
	// r := mux.NewRouter()
	//r.HandleFunc("/", handler)
	//http.Handle("/", r)
	//log.Fatal(http.ListenAndServe(":8080", nil))
}
