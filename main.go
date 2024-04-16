package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"pg-backup/internal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {

	fmt.Println("🚀 Starting pg-backup")

	// load env
	if _, err := os.Stat(".env"); err == nil {
		godotenv.Load(".env")
	} else if os.IsNotExist(err) {
		fmt.Println(".env file does not exist. Skip loading env vars.")
	} else {
		fmt.Println("error loading the environment variables from .env file. Skip loading env vars.")
	}

	dbc, err := newConn()
	if err != nil {
		fmt.Println("creating the connection error : ", err)
	}

	s, err := gocron.NewScheduler()
	if err != nil {
		fmt.Println("gocron creation error", err)
	}

	// add a job to the scheduler
	hourStr := os.Getenv("BACKUP_HOUR")
	minuteStr := os.Getenv("BACKUP_MINUTE")

	hour, err := strconv.Atoi(hourStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to convert BACKUP_HOUR to integer: %v", err))
	}

	minute, err := strconv.Atoi(minuteStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to convert BACKUP_MINUTE to integer: %v", err))
	}

	j, err := s.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(uint(hour), uint(minute), 0))),
		gocron.NewTask(
			func() {
				backupFile := fmt.Sprintf("backup_%s.sql", TimeStamp())
				dbc.Backup(backupFile)
				err := internal.CompressBackup("/backups/" + backupFile)
				if err != nil {
					fmt.Println("compression error : ", err)
					return
				}
				err = os.Remove("/backups/" + backupFile)
				if err != nil {
					fmt.Printf("error removing the backup file %s : %s\n", backupFile, err)
				}
			},
		),
	)

	if err != nil {
		fmt.Println("", err)
	}
	// each job has a unique id
	fmt.Println(j.ID())

	// start the scheduler
	s.Start()

	// create a channel to listen for OS signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// block until a signal is received
	<-c

	fmt.Println("Shutting down the backuping system.")

	// when you're done, shut it down
	err = s.Shutdown()
	if err != nil {
		fmt.Println("", err)
	}

}

// https://gist.github.com/rustyeddy/77f17f4f0fb83cc87115eb72a23f18f7
// Renmove : from the standard RFC3339 format for filename issues
func TimeStamp() string {
	ts := time.Now().UTC().Format(time.RFC3339)
	return strings.Replace(strings.Replace(ts, ":", "", -1), "-", "", -1)
}

type dbConn struct {
	host     string
	port     string
	user     string
	dbName   string
	password string
}

func (dc *dbConn) String() {
	fmt.Printf("Host: %s\nPort: %s\nDB Name: %s\nPassword: %s\n", dc.host, dc.port, dc.dbName, dc.password)
}

func (dc *dbConn) Backup(backupFile string) error {

	// Connect to the database
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dc.host, dc.port, dc.user, dc.password, dc.dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		return nil
	}
	defer db.Close()

	os.Setenv("PGPASSWORD", dc.password)

	// Run pg_dump to backup the database
	cmd := exec.Command("pg_dump", "-h", dc.host, "-p", dc.port, "-U", dc.user, "-d", dc.dbName, "-f", "/backups/"+backupFile)
	out, err := cmd.Output()
	fmt.Println(out)
	if err != nil {
		fmt.Println("Error running pg_dump command:", err)
		return nil
	}

	fmt.Println("Backup completed successfully!")
	return nil
}

func newConn() (*dbConn, error) {
	dbc := dbConn{

		host:     os.Getenv("POSTGRES_HOST"),
		port:     os.Getenv("POSTGRES_PORT"),
		user:     os.Getenv("POSTGRES_USER"),
		password: os.Getenv("PGPASSWORD"),
		dbName:   os.Getenv("POSTGRES_DB"),
	}

	fmt.Println("Host:", dbc.host)
	fmt.Println("Port:", dbc.port)
	fmt.Println("User:", dbc.user)
	fmt.Println("Password:", strings.Repeat("*", len(dbc.password)))
	fmt.Println("DB Name:", dbc.dbName)

	return &dbc, nil
}
