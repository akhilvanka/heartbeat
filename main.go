package main

import (
	"github.com/charmbracelet/huh"
	"log"
)

var (
	country string
)

func main() {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Pick a country.").
				Options(
					huh.NewOption("United States", "US"),
					huh.NewOption("Germany", "DE"),
					huh.NewOption("Brazil", "BR"),
					huh.NewOption("Canada", "CA"),
				).
				Value(&country),
		),
	)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
}
