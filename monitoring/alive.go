package monitoring

import (
	l "log"
	"time"

	c "GoWorker/config"
	d "GoWorker/datastore"
)

func StartHeartbeat() {
	d.WriteToDb("monitoring", `INSERT INTO  alive(active, lastPing, clustername) VALUES (true, NOW(), $1)
			ON CONFLICT(clustername) DO UPDATE SET active = true, lastPing = EXCLUDED.lastPing`,
		[]interface{}{c.Clustername})

	l.Println("Hearbeat is registered")
	go SendHeartbeat()
}

func Shutdown() {
	d.WriteToDb("monitoring", "UPDATE alive SET active=false WHERE clustername = $1", []interface{}{c.Clustername})
	l.Println("Shutting down: Setting alive inactive")
}

func SendHeartbeat() {

	sql := "UPDATE alive SET lastPing=NOW() WHERE clustername = $1 AND active = true"
	for {
		time.Sleep(15 * time.Second)
		d.WriteToDb("monitoring", sql, []interface{}{c.Clustername})
	}

}
