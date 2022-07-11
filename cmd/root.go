package cmd

import (
	"fmt"
	"github.com/devopsext/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"imperva_exporter/pkg/exporter"
	"net/http"
	"os"
)

var APPNAME = "IMPERVA_EXPORTER"

func envGet(s string, d interface{}) interface{} {
	return utils.EnvGet(fmt.Sprintf("%s_%s", APPNAME, s), d)
}

var exporterOptions = exporter.ImpervaExporter{
	ListenAddress: envGet("LISTEN_ADDRESS", ":9141").(string),
	MetricsPath:   envGet("METRICS_PATH", "/metrics").(string),
	ImpervaURL:    envGet("IMPERVA_MANAGEMENT_URL", "https://my.imperva.com/api/stats/v1").(string),
	ImpervaApiId:  envGet("IMPERVA_API_ID", "").(string),
	ImpervaApiKey: envGet("IMPERVA_API_KEY", "").(string),
	ImpervaSiteId: envGet("IMPERVA_SITE_ID", "").(string),
}

var rootCmd = &cobra.Command{
	Use:   "imperva_exporter",
	Short: "A small and simple exporter for getting metrics from imperva",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFormatter(&log.JSONFormatter{})
		log.SetLevel(log.InfoLevel)

		exporter, err := exporter.NewExporter(exporterOptions)
		if err != nil {
			log.Fatal(err)
		}
		prometheus.MustRegister(exporter)
		log.Debug("Listening on address " + exporterOptions.ListenAddress)
		http.Handle(exporterOptions.MetricsPath, promhttp.Handler())
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte(`<html>
             <head><title>Imperva Exporter</title></head>
             <body>
             <h1>Imperva Exporter</h1>
             <p><a href='` + exporterOptions.MetricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
			if err != nil {
				return
			}
		})
		if err := http.ListenAndServe(exporterOptions.ListenAddress, nil); err != nil {
			log.Fatal("Error starting HTTP server")
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

}

func init() {
	flags := rootCmd.PersistentFlags()

	flags.StringVar(&exporterOptions.ListenAddress, "listen-address", exporterOptions.ListenAddress, "Imperva Exporter listen address")
	flags.StringVar(&exporterOptions.MetricsPath, "metrics-path", exporterOptions.MetricsPath, "Imperva Exporter metrics path")
	flags.StringVar(&exporterOptions.ImpervaURL, "imperpva-url", exporterOptions.ImpervaURL, "Imperva management URL")
	flags.StringVar(&exporterOptions.ImpervaApiId, "imperva-api-id", exporterOptions.ImpervaApiId, "Imperva API ID")
	flags.StringVar(&exporterOptions.ImpervaApiKey, "imperva-api-key", exporterOptions.ImpervaApiKey, "Imperva API KEY")
	flags.StringVar(&exporterOptions.ImpervaSiteId, "imperva-site-id", exporterOptions.ImpervaSiteId, "Imperva SiteID")
}
