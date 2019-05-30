package main

import (
	"app/exporter"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	promAddr        = flag.String("promAddr", ":9330", "Bind address for prometheus HTTP metrics server")
	ldapAddr        = flag.String("ldapAddr", "localhost:389", "Address of OpenLDAP server")
	ldapCACrt       = flag.String("ldapCACrt", "", "Path to CA certificate for LDAPS (optional)")
	ldapUser        = flag.String("ldapUser", "", "OpenLDAP bind username (optional)")
	ldapUseStartTLS = flag.Bool("ldapUseStartTLS", false, "Use start TLS (optional)")
	ldapPass        = flag.String("ldapPass", "", "OpenLDAP bind password (optional)")
	interval        = flag.Duration("interval", 30*time.Second, "Scrape interval")
	version         = flag.Bool("version", false, "Show version and exit")
	skipVerifyTLS   = flag.Bool("skipVerifyTLS", false, "Skip TLS certs verify, default is True (optional)")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Println("Version:", exporter.GetVersion())
		os.Exit(0)
	}

	config := exporter.NewLDAPConfig()
	promConfig := exporter.NewServerConfig()
	promConfig.Address = *promAddr

	/** Parse ldap address **/
	err := config.ParseAddr(*ldapAddr)
	if err != nil {
		log.Println("Error parsing ldap address: ", err.Error())
		os.Exit(1)
	}

	/** Load Certificate if given, and panic on error **/
	if *ldapCACrt != "" {
		err = config.LoadCACert(*ldapCACrt)
		if err != nil {
			log.Println("Error loading CA certificate file: ", err.Error())
			os.Exit(1)
		} else {
			log.Println("Successfully loaded CA cert file:", *ldapCACrt)
		}
	}

	config.Username = *ldapUser
	config.Password = *ldapPass

	if *ldapUseStartTLS {
		config.UseStartTLS = true
	}

	if *skipVerifyTLS {
		config.TLSConfig.InsecureSkipVerify = true
	}

	log.Println("Starting prometheus HTTP metrics server on", *promAddr)
	go exporter.StartMetricsServer(promConfig)

	log.Printf("Starting OpenLDAP scraper for %s, StartTLS is %v, SkipTLSVerify is %v", *ldapAddr, config.UseStartTLS, config.TLSConfig.InsecureSkipVerify)
	for range time.Tick(*interval) {
		exporter.ScrapeMetrics(&config)
	}
}
