package cmd

import (
	"fmt"
	sreCommon "github.com/devopsext/sre/common"
	sreProvider "github.com/devopsext/sre/provider"
	"github.com/devopsext/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"imperva_exporter/pkg/exporter"
	"net/http"
	"os"
	"time"
)

var APPNAME = "IMPERVA_EXPORTER"
var version = "unknown"

func envGet(s string, d interface{}) interface{} {
	return utils.EnvGet(fmt.Sprintf("%s_%s", APPNAME, s), d)
}

var exporterOptions = exporter.ImpervaExporter{
	ListenAddress: envGet("LISTEN_ADDRESS", ":9141").(string),
	MetricsPath:   envGet("METRICS_PATH", "/metrics").(string),
	ImpervaFQDN:   envGet("MANAGEMENT_FQDN", "my.imperva.com").(string),
	ImpervaApiId:  envGet("API_ID", "").(string),
	ImpervaApiKey: envGet("API_KEY", "").(string),
}

var logs = sreCommon.NewLogs()
var stdoutOptions = sreProvider.StdoutOptions{
	Format:          envGet("STDOUT_FORMAT", "template").(string),
	Level:           envGet("STDOUT_LEVEL", "info").(string),
	Template:        envGet("STDOUT_TEMPLATE", "{{.time}}\t{{.level}}\t{{.file}}\t\t{{.msg}}").(string),
	TimestampFormat: envGet("STDOUT_TIMESTAMP_FORMAT", "2006-01-02T15:04:05.000000Z07:00").(string),
	TextColors:      envGet("STDOUT_TEXT_COLORS", true).(bool),
}

var stdout *sreProvider.Stdout

func realLoop(exporter *exporter.ImpervaExporter) {
	logs.Debug("MuScrap.Lock before realLoop")
	exporter.MuScrap.Lock()
	defer exporter.MuScrap.Unlock()
	logs.Debug("in the realLoop")
	exporter.LoopCollect()
	logs.Debug("out the realLoop")
	logs.Debug("MuScrap.UnLock after realLoop")
}
func scrapLoop(exporter *exporter.ImpervaExporter) {
	logs.Debug("in the loop")
	realLoop(exporter)
	for {
		logs.Debug("startsleep")
		time.Sleep(10 * time.Second)
		logs.Debug("stopsleep")
		realLoop(exporter)
	}
}

var rootCmd = &cobra.Command{
	Use:   "imperva_exporter",
	Short: "A small and simple exporter for getting metrics from imperva",
	Run: func(cmd *cobra.Command, args []string) {
		logs.Debug("cobra.Command.Run")

		exporter, err := exporter.NewExporter(exporterOptions, logs)

		if err != nil {
			logs.Panic(err)
		}
		prometheus.MustRegister(exporter)

		go scrapLoop(exporter)

		logs.Debug("Listening on address " + exporterOptions.ListenAddress)
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
			logs.Panic("Error starting HTTP server")
		}
	},
}

func Execute() {
	//var datadogEventer *sreProvider.DataDogEventer
	logs.Debug("execute")
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

}

func init() {
	flags := rootCmd.PersistentFlags()

	stdoutOptions.Version = version
	stdout = sreProvider.NewStdout(stdoutOptions)
	stdout.SetCallerOffset(2)
	logs.Register(stdout)
	logs.Debug("init")

	flags.StringVar(&exporterOptions.ListenAddress, "listen-address", exporterOptions.ListenAddress, "Imperva Exporter listen address")
	flags.StringVar(&exporterOptions.MetricsPath, "metrics-path", exporterOptions.MetricsPath, "Imperva Exporter metrics path")
	flags.StringVar(&exporterOptions.ImpervaFQDN, "imperpva-url", exporterOptions.ImpervaFQDN, "Imperva management FQDN")
	flags.StringVar(&exporterOptions.ImpervaApiId, "imperva-api-id", exporterOptions.ImpervaApiId, "Imperva API ID")
	flags.StringVar(&exporterOptions.ImpervaApiKey, "imperva-api-key", exporterOptions.ImpervaApiKey, "Imperva API KEY")
}
