package main


import (
	"context"
	"os"
	"time"
	"encoding/json"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	otext "github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/spf13/cobra"
	pb "github.com/cohenjo/waste/go/grpc/waste"
	"google.golang.org/grpc"
	"github.com/spf13/viper"
	logger "github.com/rs/zerolog/log"
)

var (
	payload    *pb.Change
	dialTarget string
	leaders string
	ghosts string
	rootCmd    = &cobra.Command{
		Use:   "wastecli",
		Short: "wastecli - Send changes to waste",
		Long: `This util implements a simple grpc client to aid testing WASTE.
				waste is expected to be called from Shepherd - more on that one later ;) `,
		Run: func(cmd *cobra.Command, args []string) {
			// create connection
			conn, err := grpc.Dial(dialTarget, grpc.WithInsecure(),
				grpc.WithUnaryInterceptor(
					otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),
				grpc.WithStreamInterceptor(
					otgrpc.OpenTracingStreamClientInterceptor(opentracing.GlobalTracer())))
			if err != nil {
				logger.Error().Err(err).Msgf("Unable to Dial to target: %s", dialTarget)
				os.Exit(1)
			}
			defer conn.Close()

			err = json.Unmarshal([]byte(leaders), &payload.Leaders)
			if err != nil {
				logger.Error().Err(err).Msgf("Failed to json leaders: %s",leaders)
			}
			err = json.Unmarshal([]byte(ghosts), &payload.Groups)
			if err != nil {
				logger.Error().Err(err).Msgf("Failed to json groups: %s",ghosts)
			}

			// create client
			client := pb.NewWasteClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			sp := opentracing.StartSpan("RunChange")
			sp.LogFields(
				log.String("dialTarget", dialTarget),
				// log.String("payload.FailureType", payload.FailureType),
				// log.String("payload.FailureDescription", payload.FailureDescription),
				// log.String("payload.FailedHost", payload.FailedHost),
				// log.Int32("payload.FailedPort", payload.FailedPort),
				// log.String("payload.FailureCluster", payload.FailureCluster),
				// log.String("payload.FailureClusterAlias", payload.FailureClusterAlias),
				// log.Int32("payload.CountReplicas", payload.CountReplicas),
				// log.Bool("payload.IsDowntimed", payload.IsDowntimed),
				// log.Bool("payload.AutoMasterRecovery", payload.AutoMasterRecovery),
				// log.Bool("payload.AutoIntermediateMasterRecovery", payload.AutoIntermediateMasterRecovery),
				// log.String("payload.OrchestratorHost", payload.OrchestratorHost),
				// log.Int32("payload.LostReplicas", payload.LostReplicas),
				// log.String("payload.ReplicaHosts", payload.ReplicaHosts),
			)
			defer sp.Finish()
			// inject the span to the context
			ctx = opentracing.ContextWithSpan(ctx, sp)
			// logger.WithSpan(sp).Info().Msg("Sending payload...")
			res, err := client.RunChange(ctx, payload)
			if err != nil {
				otext.Error.Set(sp, true)
				// logger.WithSpan(sp).Error().Err(err).Msg("PushFailure Failed.")
				logger.Error().Err(err).Msg("PushFailure Failed.")
				os.Exit(1)

			}
			// logger.WithSpan(sp).Info().Msgf("return with: %s", res.String())
			logger.Info().Msgf("return with: %s", res.String())
			return
		},
	}
)

func init() {
	
	payload = &pb.Change{}
	rootCmd.Flags().StringVarP(&dialTarget, "waste-dial-target", "t", "", "GRPC endpont that run waste.waste service, e.g: --waste-dial-target='localhost:3006'")
	rootCmd.MarkFlagRequired("waste-dial-target")
	rootCmd.Flags().StringVar(&payload.Artifact, "artifact", "", "artifact")
	rootCmd.MarkFlagRequired("artifact")
	rootCmd.Flags().StringVar(&payload.Cluster, "cluster", "", "cluster name")
	rootCmd.Flags().StringVar(&payload.Db, "db", "", "db name")
	rootCmd.Flags().StringVar(&payload.Table, "table", "", "")
	rootCmd.Flags().StringVar(&payload.Ddl, "ddl", "", "")
	rootCmd.Flags().StringVarP(&leaders, "leaders","l", "", "")
	rootCmd.Flags().StringVarP(&ghosts, "ghosts","g", "", "")
}

func main() {
	viper.Set("Debug", false)
	viper.Set("TracerEnable", true)
	viper.Set("TracerJaegerAgentAddress", "127.0.0.1:6831")
	
	// util.Setup(os.Stdout)
	// closer, _ := util.SetupTracer("wastecli")
	// logger = util.NewContextualLogger(map[string]string{"cmd": "wastecli"})

	// defer closer.Close()

	if err := rootCmd.Execute(); err != nil {

		os.Exit(1)
	}
}
