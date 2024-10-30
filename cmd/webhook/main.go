package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
)

// admission webhook which prevents server pods from being evicted

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	mux := &http.ServeMux{}
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
			return
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		s := kjson.NewSerializer(kjson.DefaultMetaFactory, scheme.Scheme, scheme.Scheme, false)

		review := &admissionv1.AdmissionReview{}
		if _, _, err = s.Decode(b, nil, review); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if review.Request == nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		req, _ := json.MarshalIndent(review, "", "  ")
		log.Print(string(req))

		review.Response = &admissionv1.AdmissionResponse{
			UID:     review.Request.UID,
			Allowed: true,
		}

		if strings.HasPrefix(review.Request.Name, "server-") {
			review.Response = &admissionv1.AdmissionResponse{
				UID:     review.Request.UID,
				Allowed: false,
				Result: &metav1.Status{
					Status:  metav1.StatusFailure,
					Message: "Cannot evict pod as it would violate the pod's disruption budget.",
					Reason:  metav1.StatusReasonTooManyRequests,
					Details: &metav1.StatusDetails{
						Causes: []metav1.StatusCause{
							{
								Type:    "DisruptionBudget",
								Message: "The disruption budget server needs 1 healthy pods and has 1 currently",
							},
						},
					},
					Code: http.StatusTooManyRequests,
				},
			}
		}

		review.Request = nil

		w.Header().Set("Content-Type", "application/json")
		s.Encode(review, w)
	})

	s := &http.Server{
		Addr:    ":8443",
		Handler: mux,
	}

	return s.ListenAndServeTLS("tls.crt", "tls.key")
}
