package datastore

import (
	"context"
	"log"
	"strings"

	c "GoWorker/config"

	"github.com/jackc/pgx/v5/pgxpool"
	//	"github.com/jackc/pgx/v5"
)

type pgConn struct {
	connStr  string
	database string
	pool     *pgxpool.Pool
}

//var metricsdb

var pgConnections map[string]*pgConn

// here we may expand it
// FUTURE TODO: put into a proper Db configuration.
func InitPgConnections() bool {
	pgConnections = make(map[string]*pgConn, 4)

	dbcred, ok := c.GetDbCredentials("DB_METRICS_MONITOR")
	if !ok {
		log.Fatal("Can't connect to metrics db - nothing is set in env")
	}
	connString := strings.Join([]string{
		"postgresql://", dbcred, "@", c.ProxyIP, ":5432/metrics"}, "")
	pgConnections["monitoring"] = &pgConn{
		connStr:  connString,
		database: "metrics",
	}
	pgConnections["monitoring"].pool = pgPoolConnection(connString)
	log.Println("Connected to postgres on ", c.ProxyIP)

	connString = strings.Join([]string{
		"postgresql://", dbcred, "@", c.Config.RegionalMetricsDb, ":5432/metrics"}, "")
	pgConnections["metrics"] = &pgConn{
		connStr:  connString,
		database: "metrics",
	}
	pgConnections["metrics"].pool = pgPoolConnection(connString)
	log.Println("Connected to postgres on ", c.Config.RegionalMetricsDb)
	return true
}

func pgPoolConnection(connString string) *pgxpool.Pool {
	//log.Println("DB CONN STRING: ", connString)
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatal(err.Error())
	}
	return pool
}

func Shutdown() *pgxpool.Conn {

	for db := range pgConnections {
		pgdb, ok := pgConnections[db]
		if !ok {
			log.Printf("Cannot get '%s' database from local hash for closing", db)
		} else {
			pgdb.pool.Close()
			log.Printf("Closed %s db", pgdb.database)
		}
	}
	return nil

}

func AcquirePg(db string) *pgxpool.Conn {
	pgdb, ok := pgConnections[db]
	if !ok {
		log.Fatalf("Cannot get '%s' database from local hash", db)
	}
	conn, err := pgdb.pool.Acquire(context.Background())
	if err != nil {
		log.Println(err.Error())
		return nil
		// send alert
	}
	return conn
}

func WriteToDb(name string, sql string, dbRows []interface{}) {

	if len(dbRows) == 0 {
		return
	}
	conn := AcquirePg(name)
	if conn == nil {
		return
	}
	defer conn.Release()
	r, err := conn.Query(context.Background(), sql, dbRows...) // XXX Exec sound more logical but there is a bug in the pgx lib...
	defer r.Close()
	if err != nil {
		// this is not nice. Panic is also not good.
		log.Println(err.Error())
	}
	return
}
