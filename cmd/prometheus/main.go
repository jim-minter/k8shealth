package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// example use of managed prometheus client

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	token, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://prometheus.monitor.azure.com/.default"},
	})
	if err != nil {
		return err
	}

	client, err := api.NewClient(api.Config{
		Address: os.Args[1],
		RoundTripper: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			req.Header.Set("Authorization", "Bearer "+token.Token)
			return api.DefaultRoundTripper.RoundTrip(req)
		}),
	})
	if err != nil {
		return err
	}

	prometheus := promv1.NewAPI(client)

	value, _, err := prometheus.Query(ctx, `up`, time.Now())
	if err != nil {
		return err
	}

	b, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = fmt.Println(string(b))
	return err
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (rt roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}
