package main

import (
	"database/sql"
	"log"
	"math"

	"github.com/douglasmg7/money"
)

type xmlProduct struct {
	Code        string `xml:"codigo,attr"`
	Brand       string `xml:"marca,attr"`
	Category    string `xml:"categoria,attr"`
	Description string `xml:"descricao,attr"`
	// Unidade           string `xml:"unidade,attr"`
	Multiplo    string `xml:"multiplo,attr"`
	DealerPrice string `xml:"preco,attr"`
	// Suggestion price to sell.
	SuggestionPrice string `xml:"precoeup,attr"`
	Weight          string `xml:"peso,attr"`
	TecDesc         string `xml:"descricao_tecnica,attr"`
	Availability    string `xml:"disponivel,attr"`
	// Ipi               string `xml:"ipi,attr"`
	Measurements string `xml:"dimensoes,attr"`
	// Abnt              string `xml:"abnt,attr"`
	// Ncm               string `xml:"ncm,attr"`
	// Origem            string `xml:"origem,attr"`
	// Ppb               string `xml:"ppb,attr"`
	// Portariappb       string `xml:"portariappb,attr"`
	// Mpdobem           string `xml:"mpdobem,attr"`
	// Dataportariappb   string `xml:"dataportariappb,attr"`
	// Icms              string `xml:"icms,attr"`
	// Reducao           string `xml:"reducao,attr"`
	// Precocomst        string `xml:"precocomst,attr"`
	// Produtocomst      string `xml:"produtocomst,attr"`
	PictureLinks string `xml:"foto,attr"`
	// DescricaoAmigavel string `xml:"descricao_amigavel,attr"`
	// CategoriaTi       string `xml:"categoria_ti,attr"`
	WarrantyTime string `xml:"tempo_garantia,attr"`
	RMAProcedure string `xml:"procedimentos_rma,attr"`
	// YoutubeLink         string `xml:"link_youtube,attr"`
	// EmpFilial         string `xml:"emp_filial,attr"`
	// Potencia          string `xml:"potencia,attr"`
}

type xmlDoc struct {
	Products []xmlProduct `xml:"produto"`
}

func (doc *xmlDoc) process() (err error) {
	// Price.
	var minPrice money.Money
	minPrice = math.MaxFloat32
	var maxPrice money.Money
	// var maxPriceCodeProduct string
	// var maxPriceDescriptionProduct string
	var priceSum money.Money
	var prodcutQtyCutByMaxPrice int
	var prodcutQtyCutByMinPrice int
	var prodcutQtyCutByCategFilter int
	mCategoryAllQtd := map[string]int{}
	mCategoryUsedQtd := map[string]int{}
	// var brand map[string]int
	// List of categories to get.
	// var available int

	var totalProductQtd int
	var usedProductQtd int

	for _, xmlProduct := range doc.Products {
		totalProductQtd++
		// List all categories.
		elem, _ := mCategoryAllQtd[xmlProduct.Category]
		mCategoryAllQtd[xmlProduct.Category] = elem + 1
		// Filter by categories.
		if !isCategorieHabToBeUsed(xmlProduct.Category) {
			prodcutQtyCutByCategFilter++
			continue
		}
		product := Product{}
		// Categories.
		product.Category = xmlProduct.Category
		// List used categories.
		elem, _ = mCategoryUsedQtd[product.Category]
		mCategoryUsedQtd[product.Category] = elem + 1

		// Price.
		var err error
		product.DealerPrice, err = money.Parse(xmlProduct.DealerPrice, ",")
		if err != nil {
			log.Printf("Could not convert dealer price, product code: %s, price: %s\n", xmlProduct.Code, xmlProduct.DealerPrice)
			continue
		}

		// Suggestion price.
		var err error
		product.SuggestionPrice, err = money.Parse(xmlProduct.SuggestionPrice, ",")
		if err != nil {
			log.Printf("Could not convert suggestion price, product code: %s, price: %s\n", xmlProduct.Code, xmlProduct.SuggestionPrice)
			continue
		}

		// Filter max price.
		if product.DealerPrice > config.FilterMaxPrice {
			prodcutQtyCutByMaxPrice++
			continue
		}
		// Filter min price.
		if product.DealerPrice < config.FilterMinPrice {
			prodcutQtyCutByMinPrice++
			continue
		}

		// Product will be used.
		usedProductQtd++
		// Code.
		product.Code = xmlProduct.Code
		// Brands.
		product.Brand = xmlProduct.Brand
		// Description.
		product.Description = xmlProduct.Description

		product.Length = 1
		product.Width = 2
		product.Height = 3
		product.Weight = 4
		product.PictureLinks = "asdf"
		product.WarrantyPeriod = 5
		product.RMAProcedure = "no-procedure"

		dbProduct := Product{}
		err = dbProduct.Find(product.Code)
		// New product.
		if err == sql.ErrNoRows {
			log.Println("Inserting:", product.Code)
			err = product.Save()
			if err != nil {
				log.Fatal(err)
			}
		} else if err != nil {
			log.Fatal(err)
		} else {
			log.Println("DealerPrice:", dbProduct.DealerPrice)
		}

		// fmt.Println("DealerPrice: ", product.DealerPrice)
		// Max price.
		if product.DealerPrice > maxPrice {
			maxPrice = product.DealerPrice
			// maxPriceCodeProduct = product.Code
			// maxPriceDescriptionProduct = product.Description
		}
		// Min price.
		if product.DealerPrice < minPrice {
			minPrice = product.DealerPrice
		}

		// Pric sum.
		priceSum += product.DealerPrice

		// fmt.Printf("[%s] - %s - R$%.2f\n", product.Category, product.Description, product.DealerPrice)
		// log.Println(product.DealerPrice)
		// log.Println()
	}
	// Average price.
	// averagePrice := priceSum.Divide(len(products))

	// log.Printf("Min price: %.2f\n", minPrice)
	// log.Printf("Max price: %.2f\n", maxPrice)
	// log.Printf("Max price code product: %s\n", maxPriceCodeProduct)
	// log.Printf("Max price desc product: %s\n", maxPriceDescriptionProduct)
	// log.Printf("Sum price: %f", priceSum)
	// log.Printf("Average price: %.4f", averagePrice)
	log.Printf("Products quantity: %d", totalProductQtd)
	log.Printf("Products quantity cut by min price(%.2f): %d", config.FilterMinPrice, prodcutQtyCutByMinPrice)
	log.Printf("Products quantity cut by max price(%.2f): %d", config.FilterMaxPrice, prodcutQtyCutByMaxPrice)
	log.Printf("Products quantity cut by categories filter: %d", prodcutQtyCutByCategFilter)
	log.Printf("Product used quantity: %d", usedProductQtd)
	log.Printf("All  Categories quantity: %d", len(mCategoryAllQtd))
	log.Printf("Used Categories quantity: %d", len(mCategoryUsedQtd))
	writeList(&mCategoryUsedQtd, "list/categUse.list")
	writeList(&mCategoryAllQtd, "list/categAll.list")
	return err
}