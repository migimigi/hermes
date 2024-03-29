package pricing

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"testing"
)

func TestFetchRedshift(t *testing.T) {
	p, err := Fetch(Redshift, "ap-northeast-1")
	if err != nil {
		t.Error(err)
	}

	price := make([]Price, 0)
	for _, v := range p {
		price = append(price, v)
	}

	if len(price) < 1 {
		t.Fail()
	}
}

func TestDiscountRate(t *testing.T) {
	price, err := Deserialize("/var/tmp/hermes", []string{"ap-northeast-1"})
	if err != nil {
		t.Errorf("desirialize: %v", err)
	}

	sort.SliceStable(price, func(i, j int) bool { return price[i].DiscountRate() > price[j].DiscountRate() })

	content := make([]string, 0)
	for _, p := range price {
		line := fmt.Sprintf("%.2f, %v, %v, %v, %v, %v, %v, %v, %v, %f, %f, %f, %v, %v, %v, %v, %v, %v, %v, %v　\n",
			p.DiscountRate(),
			p.Version,
			p.SKU,
			p.OfferTermCode,
			p.Region,
			p.InstanceType,
			p.UsageType,
			p.LeaseContractLength,
			p.PurchaseOption,
			p.OnDemand,
			p.ReservedQuantity,
			p.ReservedHrs,
			p.Tenancy,
			p.PreInstalled,
			p.Operation,
			p.OperatingSystem,
			p.CacheEngine,
			p.DatabaseEngine,
			p.OfferingClass,
			p.NormalizationSizeFactor,
		)

		content = append(content, line)
	}

	var str string
	for i := range content {
		str = str + content[i]
	}

	bytes := []byte(str)
	if err := ioutil.WriteFile("/var/tmp/hermes/ri_dc_rate.csv", bytes, os.ModePerm); err != nil {
		t.Errorf("write file: %v", err)
	}
}

func TestBreakEvenPoint(t *testing.T) {
	cases := []struct {
		Price Price
		Point int
	}{
		{
			Price{
				Region:                  "ap-northeast-1",
				UsageType:               "APN1-BoxUsage:c4.large",
				Tenancy:                 "Shared",
				PreInstalled:            "NA",
				OperatingSystem:         "Linux",
				OfferingClass:           "standard",
				LeaseContractLength:     "1yr",
				PurchaseOption:          "All Upfront",
				OnDemand:                0.126,
				ReservedQuantity:        738,
				ReservedHrs:             0,
				NormalizationSizeFactor: "4",
			},
			9,
		},
		{
			Price{
				Region:                  "ap-northeast-1",
				UsageType:               "APN1-BoxUsage:c4.large",
				Tenancy:                 "Shared",
				PreInstalled:            "NA",
				OperatingSystem:         "Linux",
				OfferingClass:           "standard",
				LeaseContractLength:     "1yr",
				PurchaseOption:          "Partial Upfront",
				OnDemand:                0.126,
				ReservedQuantity:        377,
				ReservedHrs:             0.043,
				NormalizationSizeFactor: "4",
			},
			7,
		},
		{
			Price{
				Region:                  "ap-northeast-1",
				UsageType:               "APN1-BoxUsage:c4.large",
				Tenancy:                 "Shared",
				PreInstalled:            "NA",
				OperatingSystem:         "Linux",
				OfferingClass:           "standard",
				LeaseContractLength:     "1yr",
				PurchaseOption:          "No Upfront",
				OnDemand:                0.126,
				ReservedQuantity:        0,
				ReservedHrs:             0.09,
				NormalizationSizeFactor: "4",
			},
			1,
		},
	}

	for _, tt := range cases {
		if tt.Price.BreakEvenPoint() != tt.Point {
			t.Errorf("expected: %v, actual: %v", tt.Point, tt.Price.BreakEvenPoint())
		}
	}
}
