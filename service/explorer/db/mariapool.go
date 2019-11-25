package mariadbpool

import (

	// mysql driver

	"errors"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"

	// gorm mysql driver
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

const (
	maxIdle       = 20
	maxConnection = 200
)

var cp *connectionPool

func init() {
	cp = &connectionPool{
		actives:       map[string]*DB{},
		idles:         []*DB{},
		maxIdle:       maxIdle,
		maxConnection: maxConnection,
		waitIdleConn:  make(chan bool, maxConnection),
	}
}

// Connect DB
func Connect() *DB {
	db, err := cp.GetIdleConn()
	if err == nil {
		return db
	}
	if err != ErrNoIdleConn {
		panic(err)
	}
	if len(cp.idles)+len(cp.actives) < cp.maxConnection {
		return cp.NewConnection()
	}

	pingTicker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-pingTicker.C:
			db, err := cp.GetIdleConn()
			if err == nil {
				return db
			}
			log.Println("no idle conn")
			pingTicker = time.NewTicker(100 * time.Millisecond)
		case <-cp.waitIdleConn:
			db, err := cp.GetIdleConn()
			if err == nil {
				for range cp.waitIdleConn {
				}
				return db
			}
			pingTicker = time.NewTicker(100 * time.Millisecond)
		}
	}
}

var (
	ErrNoIdleConn = errors.New("No Idle Connection")
)

type connectionPool struct {
	actives       map[string]*DB
	idles         []*DB
	maxIdle       int
	maxConnection int
	waitIdleConn  chan bool
	lock          sync.Mutex
}

func (cp *connectionPool) NewConnection() *DB {
	db := New()
	cp.actives[db.uuid] = db
	return db
}

func (cp *connectionPool) GetIdleConn() (*DB, error) {
	cp.lock.Lock()
	defer cp.lock.Unlock()

	if len(cp.idles) == 0 {
		return nil, ErrNoIdleConn
	}

	db := cp.idles[0]
	cp.idles = cp.idles[1:]

	m := struct{}{}
	if err := db.Raw(`SELECT * FROM INFORMATION_SCHEMA.TABLES limit 1`).Find(&m).Error; err != nil {
		db = New()
	}
	cp.actives[db.uuid] = db

	return db, nil
}

func (s *DB) Close() error {
	delete(s.cp.actives, s.uuid)
	var err error
	if len(cp.idles) >= maxIdle {
		err = s.DB.Close()
	} else {
		cp.idles = append(cp.idles, s)
	}
	go func() {
		cp.waitIdleConn <- true
	}()
	return err
}

type DB struct {
	*gorm.DB
	cp   *connectionPool
	uuid string
}

func New() *DB {
	db, err := gorm.Open("mysql", "portal_admin:nhcq3t40n9y32t4f0n8@tcp(49.247.203.232)/f_portal?charset=utf8mb4&parseTime=True&loc=Local")
	//db, err := gorm.Open("mysql", "portal_admin:nhcq3t40n9y32t4f0n8@tcp(127.0.0.1)/f_portal?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		// panic("failed to connect database")
		panic(err)
	}
	db.DB().SetMaxIdleConns(maxIdle)
	db.DB().SetMaxOpenConns(maxConnection * 2)
	db.DB().SetConnMaxLifetime(time.Hour)

	DB := &DB{
		db,
		cp,
		uuid.NewV1().String(),
	}

	return DB
}
