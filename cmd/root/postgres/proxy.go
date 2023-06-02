package postgres

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/nais/cli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy [app-name] [flags]",
	Short: "Create a proxy to a Postgres instance",
	Long:  `Update IAM policies by giving your user the a timed sql.cloudsql.instanceUser role, then start a proxy to the instance.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		ctx := context.Background()
		appName := args[0]
		namespace := viper.GetString(cmd.NamespaceFlag)
		context := viper.GetString(cmd.ContextFlag)
		port := viper.GetString(cmd.PortFlag)
		host := viper.GetString(cmd.HostFlag)
		databaseName := viper.GetString(cmd.DatabaseFlag)

		dbInfo, err := NewDBInfo(appName, namespace, context, databaseName)
		if err != nil {
			return err
		}

		projectID, err := dbInfo.ProjectID(ctx)
		if err != nil {
			return err
		}

		connectionName, err := dbInfo.ConnectionName(ctx)
		if err != nil {
			return err
		}
		connectionInfo, err := dbInfo.DBConnection(ctx)
		if err != nil {
			return err
		}
		email, err := currentEmail(ctx)
		if err != nil {
			return err
		}

		address := fmt.Sprintf("%v:%v", host, port)

		fmt.Printf("Starting proxy on %v\n", address)
		fmt.Println("If you are using psql, you can connect to the database by running:")
		fmt.Printf("psql -h %v -p %v -U %v %v\n", host, port, email, connectionInfo.dbName)
		fmt.Println()
		fmt.Println("If you are using a JDBC client, you can connect to the database by using the following connection string:")
		fmt.Printf("Connection URL: jdbc:postgresql://%v/%v?user=%v\n", address, connectionInfo.dbName, email)
		fmt.Println()
		fmt.Println("If you get asked for a password, you can leave it blank. If that doesn't work, try running 'nais postgres grant", appName+"' again.")

		return runProxy(ctx, projectID, connectionName, address, make(chan int, 1))
	},
}

func runProxy(ctx context.Context, projectID, connectionName, address string, port chan int) error {
	if err := grantUserAccess(ctx, projectID, "roles/cloudsql.instanceUser", 1*time.Hour); err != nil {
		return err
	}

	opts := []cloudsqlconn.Option{
		cloudsqlconn.WithIAMAuthN(),
	}
	d, err := cloudsqlconn.NewDialer(context.Background(), opts...)
	if err != nil {
		return fmt.Errorf("failed to create dialer: %w", err)
	}

	if err := d.Warmup(ctx, connectionName); err != nil {
		return fmt.Errorf("failed to warmup connection: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on TCP address: %w", err)
	}

	fmt.Println("Listening on", listener.Addr().String())

	port <- listener.Addr().(*net.TCPAddr).Port

	go func() {
		<-ctx.Done()
		// TODO: Make this not panic listener.Accept()
		if err := listener.Close(); err != nil {
			log.Println("error closing listener", err)
		}
	}()

	wg := sync.WaitGroup{}
OUTER:
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				break OUTER
			default:
			}
			log.Println("error accepting connection", err)
			continue
		}
		log.Println("New connection", conn.RemoteAddr())
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			go func() {
				<-ctx.Done()
				if err := conn.Close(); err != nil {
					log.Println("error closing connection", err)
				}
			}()

			conn2, err := d.Dial(ctx, connectionName)
			if err != nil {
				log.Println("error dialing connection", err)
				return
			}
			defer conn2.Close()

			closer := make(chan struct{}, 2)
			go copy(closer, conn2, conn)
			go copy(closer, conn, conn2)
			<-closer
			log.Println("Connection complete", conn.RemoteAddr())
		}()
	}

	fmt.Println("Waiting for connections to close")
	wg.Wait()

	return nil
}

func copy(closer chan struct{}, dst io.Writer, src io.Reader) {
	_, _ = io.Copy(dst, src)
	closer <- struct{}{} // connection is closed, send signal to stop proxy
}
