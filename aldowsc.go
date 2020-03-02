package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/douglasmg7/aldoutil"
	"github.com/douglasmg7/currency"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	// "golang.org/x/net/html/charset"
	// "golang.org/x/text/encoding/charmap"
	// "code.google.com/p/go-charset/charset"
	// _ "code.google.com/p/go-charset/data"
)

var db *sql.DB
var dbAldo *sqlx.DB

// Paths.
var appPath, logPath, dbPath, xmlPath string

// Files.
var logFile, dbAldoFile string

// Min and max price filter.
var maxPriceFilter, minPriceFilter currency.Currency

// Development mode.
var dev bool

// Categories selected to use, key is the name for category.
var selectedCategories map[string]aldoutil.Category

func init() {
	// Path.
	zunkaPath := os.Getenv("ZUNKAPATH")
	if zunkaPath == "" {
		panic("ZUNKAPATH not defined.")
	}
	logPath := path.Join(zunkaPath, "log", "aldo")
	xmlPath = path.Join(zunkaPath, "xml")
	// Create path.
	os.MkdirAll(logPath, os.ModePerm)
	os.MkdirAll(xmlPath, os.ModePerm)

	// Aldo db.
	dbAldoFile = os.Getenv("ZUNKA_ALDOWSC_DB")
	if dbAldoFile == "" {
		panic("ZUNKA_ALDOWSC_DB not defined.")
	}
	dbAldoFile = path.Join(zunkaPath, "db", dbAldoFile)

	// Log file.
	logFile, err := os.OpenFile(path.Join(logPath, "aldowsc.log"), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	// Filters.
	// minPriceFilter, err = currency.Parse("1.400,00", ",")
	minPriceFilter, err = currency.Parse("1050,00", ",")
	if err != nil {
		panic(err)
	}
	// maxPriceFilter, err = currency.Parse("100.000,00", ",")
	maxPriceFilter, err = currency.Parse("100.000.000,00", ",")
	if err != nil {
		panic(err)
	}

	// Log configuration.
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	log.SetFlags(log.Ldate | log.Lmicroseconds)
	// log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	// log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	// log.SetFlags(log.LstdFlags | log.Ldate | log.Lshortfile)
	// log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	// Run mode.
	mode := "production"
	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "dev") {
		dev = true
		mode = "development"
	}
	// Log start.
	log.Printf("*** Starting aldowsc in %v mode (version %s) ***\n", mode, version)
	// log.Printf("*** Starting aldowsc in %v mode (version %s) ***\n", mode, "1")
}

func main() {
	var err error

	// Db start.
	db, err = sql.Open("sqlite3", dbAldoFile)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Init db Aldo.
	dbAldo = sqlx.MustConnect("sqlite3", dbAldoFile)

	// Getting selected categories.
	log.Println("Reading selected categories from db...")
	selectedCategoriesArray := []aldoutil.Category{}
	err = dbAldo.Select(&selectedCategoriesArray, "SELECT * FROM category where selected=true")
	if err != nil {
		log.Fatalln("Getting categories from db:", err)
	}
	selectedCategories = map[string]aldoutil.Category{}
	for _, category := range selectedCategoriesArray {
		selectedCategories[category.Name] = category
	}
	// log.Println("selectedCategories:", selectedCategories)

	// Remove no more selected products.
	rmProductsNotSel()

	// Remove products with price out of range.
	rmProductsPriceOutOfRange()

	// Load xml file.
	log.Println("Loading and decoding xml file...")
	// timer := time.Now()
	aldoXMLDoc := xmlDoc{}
	decoder := xml.NewDecoder(os.Stdin)

	// Decoding xml file.
	timer := time.Now()
	// decoder.CharsetReader = charset.NewReaderLabel
	// decoder.CharsetReader = makeCharsetReader
	// decoder.CharsetReader = charset.NewReader
	err = decoder.Decode(&aldoXMLDoc)
	if err != nil {
		log.Fatalln("Error decoding xml file:", err)
	}
	log.Printf("Time loading and decoding xml file: %fs", time.Since(timer).Seconds())
	// log.Printf("codigo: %q\n", aldoXMLDoc.Products[0].Code)
	// log.Printf("descricao_tecnica: %v\n", aldoXMLDoc.Products[0].TechnicalDescription)

	// Not process zero products, to not remove current product.
	if len(aldoXMLDoc.Products) == 0 {
		log.Println("*** XML returned zero products ***")
		log.Printf("Finish\n\n")
		return
	}

	// Processing products.
	log.Println("Processing products...")
	timer = time.Now()
	err = aldoXMLDoc.process()
	log.Printf("Time processing products: %fs", time.Since(timer).Seconds())
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Finish\n\n")
}

/**************************************************************************************************
* Util.
**************************************************************************************************/

// func makeCharsetReader(charset string, input io.Reader) (io.Reader, error) {
// if charset == "iso-8859-1" {
// // Windows-1252 is a superset of ISO-8859-1, so should do here
// return charmap.Windows1252.NewDecoder().Reader(input), nil
// }
// return nil, fmt.Errorf("Unknown charset: %s", charset)

// }

func updateDBCategories(m *map[string]int) {
	// Each category.
	for category, quantity := range *m {
		text := strings.ToLower(category)
		name := strings.ReplaceAll(text, " ", "")
		// fmt.Printf("name: %s\n", name)
		// fmt.Printf("text: %s\n", text)
		// fmt.Printf("productsQty: %v\n", quantity)
		// fmt.Printf("selected: %v\n", false)
		dbAldo.MustExec(fmt.Sprintf("INSERT INTO category(name, text, products_qty, selected) VALUES(\"%s\", \"%s\", %v, %v) ON CONFLICT(name) DO UPDATE SET products_qty=excluded.products_qty", name, text, quantity, false))
	}
}

// Remove no more selected products from db.
func rmProductsNotSel() {
	// Get distinct categories from products on db.
	dbCategs := []string{}
	stmGet := "SELECT distinct category FROM product"
	err := dbAldo.Select(&dbCategs, stmGet)
	checkFatalSQLError(err, stmGet)
	// if err != nil {
	// log.Fatal(fmt.Errorf("Get distinct categories from db. %v", err))
	// }
	// Categories to be removed.
	categToRem := []string{}
	for _, dbCateg := range dbCategs {
		// fmt.Println("dbCateg:", strings.ReplaceAll(strings.ToLower(dbCateg), " ", ""))
		// fmt.Println("selectedCategories:", selectedCategories)
		if _, ok := selectedCategories[strings.ReplaceAll(strings.ToLower(dbCateg), " ", "")]; !ok {
			categToRem = append(categToRem, `"`+dbCateg+`"`)
		}
	}
	if len(categToRem) > 0 {
		log.Printf("Removing no more selected categorie(s): %s.", strings.Join(categToRem, ", "))
	}

	// Copy products to remove to history.
	tx := dbAldo.MustBegin()
	// tx.MustExec(fmt.Sprintf("INSERT INTO product_history SELECT * FROM product WHERE category IN (%s)", strings.Join(categToRem, ",")))
	stmInsert := fmt.Sprintf("INSERT INTO product_history SELECT * FROM product WHERE category IN (%s)", strings.Join(categToRem, ","))
	_, err = tx.Exec(stmInsert)
	checkFatalSQLError(err, stmInsert)
	// if err != nil {
	// log.Fatal(fmt.Errorf("[ERROR] Inserting into product_history. stm: %s. %v", stmInsert, err))
	// }
	// Delete copied products.
	// tx.MustExec(fmt.Sprintf("DELETE FROM product WHERE category IN (%s)", strings.Join(categToRem, ",")))
	stmRemove := fmt.Sprintf("DELETE FROM product WHERE category IN (%s)", strings.Join(categToRem, ","))
	_, err = tx.Exec(stmRemove)
	checkFatalSQLError(err, stmRemove)
	// if err != nil {
	// log.Fatal(fmt.Errorf("Deleting product. stm: %s. %v", stmRemove, err))
	// }
	err = tx.Commit()
	checkFatalSQLError(err, stmInsert+"\n"+stmRemove)
	// if err != nil {
	// log.Fatal(fmt.Errorf("Committing: %v\n%s\n%s", err, stmInsert, stmRemove))
	// }
}

// Remove products with price out of defined range.
func rmProductsPriceOutOfRange() {
	// Copy products to remove to history.
	tx := dbAldo.MustBegin()
	// tx.MustExec(fmt.Sprintf("INSERT INTO product_history SELECT * FROM product WHERE dealer_price NOT BETWEEN (%d) AND (%d)", minPriceFilter, maxPriceFilter))
	stmInsert := fmt.Sprintf("INSERT INTO product_history SELECT * FROM product WHERE dealer_price NOT BETWEEN (%d) AND (%d)", minPriceFilter, maxPriceFilter)
	_, err := tx.Exec(stmInsert)
	checkFatalSQLError(err, stmInsert)
	// if err != nil {
	// log.Fatal(fmt.Errorf("Inserting into product_history. stm: %s. %v", stmInsert, err))
	// }
	// Delete copied products.
	// result := tx.MustExec(fmt.Sprintf("DELETE FROM product WHERE dealer_price NOT BETWEEN (%d) AND (%d)", minPriceFilter, maxPriceFilter))
	stmRemove := fmt.Sprintf(fmt.Sprintf("DELETE FROM product WHERE dealer_price NOT BETWEEN (%d) AND (%d)", minPriceFilter, maxPriceFilter))
	result, err := tx.Exec(stmRemove)
	checkFatalSQLError(err, stmRemove)
	// if err != nil {
	// log.Fatal(fmt.Errorf("Removing product. stm: %s. %v", stmRemove, err))
	// }
	err = tx.Commit()
	checkFatalSQLError(err, stmInsert+"\n"+stmRemove)
	// if err != nil {
	// log.Fatal(fmt.Errorf("Committing: %v\n%s\n%s", err, stmInsert, stmRemove))
	// }
	rowsAffected, err := result.RowsAffected()
	checkFatalSQLError(err, stmInsert+"\n"+stmRemove)
	// if err != nil {
	// log.Fatal(fmt.Errorf("Getting rows affected. %v", err))
	// }
	if rowsAffected > 0 {
		log.Printf("Removed %v product(s) with price out of range", rowsAffected)
	}
}

/**************************************************************************************************
* ERROS
**************************************************************************************************/
func checkError(err error) bool {
	if err != nil {
		// notice that we're using 1, so it will actually log where
		// the error happened, 0 = this function, we don't want that.
		function, file, line, _ := runtime.Caller(1)
		log.Printf("[error] [%s] [%s:%d] %v", filepath.Base(file), runtime.FuncForPC(function).Name(), line, err)
		return true
	}
	return false
}

func checkFatalError(err error) {
	if err != nil {
		// notice that we're using 1, so it will actually log where
		// the error happened, 0 = this function, we don't want that.
		function, file, line, _ := runtime.Caller(1)
		log.Fatalf("[error] [%s] [%s:%d] %v", filepath.Base(file), runtime.FuncForPC(function).Name(), line, err)
	}
}

func checkFatalSQLError(err error, stm string) {
	if err != nil {
		// notice that we're using 1, so it will actually log where
		// the error happened, 0 = this function, we don't want that.
		function, file, line, _ := runtime.Caller(1)
		log.Fatalf("[error] [%s] [%s:%d] %v\n%s", filepath.Base(file), runtime.FuncForPC(function).Name(), line, err, stm)
	}
}
