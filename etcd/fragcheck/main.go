package main

import (
	crand "crypto/rand"
	"errors"
	"fmt"
	mrand "math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	globalCfg = globalConfig{}
)

type dbSizeEntry struct {
	dbSize      int64
	dbSizeInUse int64
	percent     float64 // dbSizeInUse/dbSize
}

func newFragCheckCommand() *cobra.Command {
	fragCheckCmd := &cobra.Command{
		Use:   "fragcheck",
		Short: "A simple command line tool to analyze etcd fragmentation",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return viper.BindPFlags(cmd.Flags())
		},
		Run: fragCheckCommandFunc,
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("ETCD_DEFRAG")
	viper.AutomaticEnv()
	setDefaults()

	// Manually splitting, because GetStringSlice has inconsistent behavior for splitting command line flags and environment variables
	// https://github.com/spf13/viper/issues/380
	fragCheckCmd.Flags().StringSliceVar(&globalCfg.endpoints, "endpoints", strings.Split(viper.GetString("endpoints"), ","), "comma separated etcd endpoints")

	fragCheckCmd.Flags().BoolVar(&globalCfg.useClusterEndpoints, "cluster", viper.GetBool("cluster"), "use all endpoints from the cluster member list")

	fragCheckCmd.Flags().DurationVar(&globalCfg.dialTimeout, "dial-timeout", viper.GetDuration("dial-timeout"), "dial timeout for client connections")
	fragCheckCmd.Flags().DurationVar(&globalCfg.commandTimeout, "command-timeout", viper.GetDuration("command-timeout"), "command timeout (excluding dial timeout)")
	fragCheckCmd.Flags().DurationVar(&globalCfg.keepAliveTime, "keepalive-time", viper.GetDuration("keepalive-time"), "keepalive time for client connections")
	fragCheckCmd.Flags().DurationVar(&globalCfg.keepAliveTimeout, "keepalive-timeout", viper.GetDuration("keepalive-timeout"), "keepalive timeout for client connections")

	fragCheckCmd.Flags().BoolVar(&globalCfg.insecure, "insecure-transport", viper.GetBool("insecure-transport"), "disable transport security for client connections")

	fragCheckCmd.Flags().BoolVar(&globalCfg.insecureSkepVerify, "insecure-skip-tls-verify", viper.GetBool("insecure-skip-tls-verify"), "skip server certificate verification (CAUTION: this option should be enabled only for testing purposes)")
	fragCheckCmd.Flags().StringVar(&globalCfg.certFile, "cert", viper.GetString("cert"), "identify secure client using this TLS certificate file")
	fragCheckCmd.Flags().StringVar(&globalCfg.keyFile, "key", viper.GetString("key"), "identify secure client using this TLS key file")
	fragCheckCmd.Flags().StringVar(&globalCfg.caFile, "cacert", viper.GetString("cacert"), "verify certificates of TLS-enabled secure servers using this CA bundle")

	fragCheckCmd.Flags().StringVar(&globalCfg.username, "user", viper.GetString("user"), "username[:password] for authentication (prompt if password is not supplied)")
	fragCheckCmd.Flags().StringVar(&globalCfg.password, "password", viper.GetString("password"), "password for authentication (if this option is used, --user option shouldn't include password)")

	fragCheckCmd.Flags().StringVarP(&globalCfg.dnsDomain, "discovery-srv", "d", viper.GetString("discovery-srv"), "domain name to query for SRV records describing cluster endpoints")
	fragCheckCmd.Flags().StringVarP(&globalCfg.dnsService, "discovery-srv-name", "", viper.GetString("discovery-srv-name"), "service name to query when using DNS discovery")
	fragCheckCmd.Flags().BoolVar(&globalCfg.insecureDiscovery, "insecure-discovery", viper.GetBool("insecure-discovery"), "accept insecure SRV records describing cluster endpoints")

	fragCheckCmd.Flags().DurationVar(&globalCfg.compactInterval, "compact-interval", viper.GetDuration("compact-interval"), "the interval to perform compaction")
	fragCheckCmd.Flags().DurationVar(&globalCfg.testDuration, "test-duration", viper.GetDuration("test-duration"), "how long the test should run")
	fragCheckCmd.Flags().IntVar(&globalCfg.keyCount, "key-count", viper.GetInt("key-count"), "key count")
	fragCheckCmd.Flags().IntVar(&globalCfg.minValSize, "min-value-size", viper.GetInt("min-value-size"), "min value size")
	fragCheckCmd.Flags().IntVar(&globalCfg.maxValSize, "max-value-size", viper.GetInt("max-value-size"), "max value size")

	fragCheckCmd.Flags().BoolVar(&globalCfg.printVersion, "version", viper.GetBool("version"), "print the version and exit")

	return fragCheckCmd
}

func setDefaults() {
	viper.SetDefault("endpoints", "127.0.0.1:2379")
	viper.SetDefault("cluster", false)
	viper.SetDefault("dial-timeout", 2*time.Second)
	viper.SetDefault("command-timeout", 30*time.Second)
	viper.SetDefault("keepalive-time", 2*time.Second)
	viper.SetDefault("keepalive-timeout", 6*time.Second)
	viper.SetDefault("insecure-transport", true)
	viper.SetDefault("insecure-skip-tls-verify", false)
	viper.SetDefault("cert", "")
	viper.SetDefault("key", "")
	viper.SetDefault("cacert", "")
	viper.SetDefault("user", "")
	viper.SetDefault("password", "")
	viper.SetDefault("discovery-srv", "")
	viper.SetDefault("discovery-srv-name", "")
	viper.SetDefault("insecure-discovery", true)
	viper.SetDefault("compact-interval", 30*time.Second)
	viper.SetDefault("test-duration", 5*time.Minute)
	viper.SetDefault("key-count", 500)
	viper.SetDefault("min-value-size", 100)
	viper.SetDefault("max-value-size", 120)
	viper.SetDefault("version", false)
}

func main() {
	fragCheckCmd := newFragCheckCommand()
	if err := fragCheckCmd.Execute(); err != nil {
		if fragCheckCmd.SilenceErrors {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		} else {
			os.Exit(1)
		}
	}
}

func printVersion(printVersion bool) {
	if printVersion {
		fmt.Printf("fragcheck Version: %s\n", Version)
		fmt.Printf("Git SHA: %s\n", GitSHA)
		fmt.Printf("Go Version: %s\n", runtime.Version())
		fmt.Printf("Go OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}
}

func fragCheckCommandFunc(cmd *cobra.Command, args []string) {
	printVersion(globalCfg.printVersion)

	// key is endpoint, value is a slice of dbSizeEntry
	dbSizeReport := map[string][]dbSizeEntry{}

	eps, err := endpoints(globalCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get endpoints: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Endpoints: %v\n", eps)

	fmt.Println("Validating configuration.")
	if err := validateConfig(cmd, globalCfg); err != nil {
		fmt.Fprintf(os.Stderr, "Validating configuration failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Getting the initial members status")
	statusList, err := membersStatus(globalCfg, eps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get members status: %v\n", err)
		os.Exit(1)
	}

	// record the very initial db size usage
	recordDBSizeUsage(dbSizeReport, statusList)

	compactTicker := time.NewTicker(globalCfg.compactInterval)
	testTicker := time.NewTicker(globalCfg.testDuration)

	cnt := 0
	active := true
	for active {
		select {
		case <-testTicker.C:
			active = false
		case <-compactTicker.C:
			fmt.Printf("wrote key: %d\n", cnt)
			cnt = 0

			if err := compact(globalCfg, 0, eps); err != nil {
				fmt.Printf("compaction failed: %v\n", err)
				os.Exit(1)
			}
			if err := ensureMembersSynced(globalCfg, eps); err != nil {
				fmt.Printf("ensureMembersSynced failed: %v\n", err)
				os.Exit(1)
			}

			statusList, err := membersStatus(globalCfg, eps)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get members status: %v\n", err)
				os.Exit(1)
			}

			// record the very initial db size usage
			recordDBSizeUsage(dbSizeReport, statusList)
		default:
			writeData(globalCfg, eps, getKey(globalCfg), getValue(globalCfg))
			cnt++
		}
	}

	compactTicker.Stop()
	testTicker.Stop()

	fmt.Println("Done!")
}

func validateConfig(cmd *cobra.Command, gcfg globalConfig) error {
	if gcfg.certFile == "" && cmd.Flags().Changed("cert") {
		return errors.New("empty string is passed to --cert option")
	}

	if gcfg.keyFile == "" && cmd.Flags().Changed("key") {
		return errors.New("empty string is passed to --key option")
	}

	if gcfg.caFile == "" && cmd.Flags().Changed("cacert") {
		return errors.New("empty string is passed to --cacert option")
	}

	fmt.Printf("%+v\n", gcfg)

	return nil
}

func recordDBSizeUsage(report map[string][]dbSizeEntry, statusList []epStatus) {
	for _, status := range statusList {
		entry := dbSizeEntry{
			dbSize:      status.Resp.DbSize,
			dbSizeInUse: status.Resp.DbSizeInUse,
			percent:     float64(status.Resp.DbSizeInUse) / float64(status.Resp.DbSize),
		}
		if entries, ok := report[status.Ep]; ok {
			entries = append(entries, entry)
			report[status.Ep] = entries
		} else {
			report[status.Ep] = []dbSizeEntry{entry}
		}

		fmt.Printf("Endpoint: %s, dbSize: %d, dbSizeInUse: %d, used percent: %.2f\n", status.Ep, entry.dbSize, entry.dbSizeInUse, entry.percent)
	}
	fmt.Println()
}

func getKey(gcfg globalConfig) string {
	return fmt.Sprintf("key-%d", mrand.Intn(gcfg.keyCount))
}

func getValue(gcfg globalConfig) string {
	valueBytes := randomIntInRange(gcfg.minValSize, gcfg.maxValSize)
	v := make([]byte, valueBytes)
	if _, err := crand.Read(v); err != nil {
		fmt.Printf("Failed to generate value: %v\n", err)
		os.Exit(1)
	}
	return string(v)
}

func randomIntInRange(min, max int) int {
	return mrand.Intn(max-min) + min
}
