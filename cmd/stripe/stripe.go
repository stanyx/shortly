package main

import (
	"flag"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/plan"
	"github.com/stripe/stripe-go/product"

	"shortly/config"
	"shortly/storage"
)

// stripe billing products and plan creation
func main() {

	var configPath string

	flag.StringVar(&configPath, "config", "./config/config.yaml", "stripe key")
	flag.Parse()

	appConfig, err := config.ReadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	db, err := storage.StartDB(storage.GetConnString(appConfig.Database))
	if err != nil {
		log.Fatal(err)
	}

	stripe.Key = appConfig.Billing.Payment.Key

	standartPlanParams := &stripe.ProductParams{
		Name:        stripe.String("standart_billing_plan"),
		Type:        stripe.String(string(stripe.ProductTypeService)),
		Description: stripe.String("Standart billing plan"),
	}

	standartPlanProduct, err := product.New(standartPlanParams)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("standart plan product params created:\n%+v", standartPlanProduct)

	_, err = db.Exec("update billing_plans set stripe_id = $1 where id = 2", standartPlanProduct.ID)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("insert into stripe_products (stripe_id, name, created) values ($1, $2, $3)",
		standartPlanProduct.ID, standartPlanProduct.Name, standartPlanProduct.Created)
	if err != nil {
		log.Fatal(err)
	}

	proPlanParams := &stripe.ProductParams{
		Name:        stripe.String("pro_billing_plan"),
		Type:        stripe.String(string(stripe.ProductTypeService)),
		Description: stripe.String("Pro billing plan"),
	}

	proPlanProduct, err := product.New(proPlanParams)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("update billing_plans set stripe_id = $1 where id = 3", proPlanProduct.ID)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("insert into stripe_products (stripe_id, name, created) values ($1, $2, $3)",
		proPlanProduct.ID, proPlanProduct.Name, proPlanProduct.Created)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("standart plan product params created:\n%+v", proPlanProduct)

	plStandartParams := &stripe.PlanParams{
		Nickname:  stripe.String("Standard Monthly"),
		ProductID: stripe.String(standartPlanProduct.ID),
		Amount:    stripe.Int64(2900),
		Currency:  stripe.String(string(stripe.CurrencyUSD)),
		Interval:  stripe.String(string(stripe.PlanIntervalMonth)),
		UsageType: stripe.String(string(stripe.PlanUsageTypeMetered)),
	}

	plStandart, err := plan.New(plStandartParams)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("insert into stripe_plans (stripe_id, name, created) values ($1, $2, $3)",
		plStandart.ID, plStandart.Nickname, plStandart.Created)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("standart plan created", plStandart)

	plProParams := &stripe.PlanParams{
		Nickname:  stripe.String("Pro Monthly"),
		ProductID: stripe.String(proPlanProduct.ID),
		Amount:    stripe.Int64(29900),
		Currency:  stripe.String(string(stripe.CurrencyUSD)),
		Interval:  stripe.String(string(stripe.PlanIntervalMonth)),
		UsageType: stripe.String(string(stripe.PlanUsageTypeMetered)),
	}

	plPro, _ := plan.New(plProParams)

	_, err = db.Exec("insert into stripe_plans (stripe_id, name, created) values ($1, $2, $3)",
		plPro.ID, plPro.Nickname, plPro.Created)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("pro plan created", plPro)

}
