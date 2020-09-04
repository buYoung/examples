package global

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"log"
	"strings"
	"time"
)

type DB struct {
	Db         *sql.DB
	Conn       *sql.Conn
	regenstate chan bool
}

var (
	dbctx = context.Background()
)

// dbdriver : mysql을 사용하면 mysql mariadb도 기반은 mysql이므로 mysql을 적으면됨!
//            postgres를 사용하면 postgres
// conncmd : db연결할때 사용하는 연결 명령어 mysql 기준 id:pw@tcp(ip:port)/dbname?option
//           option ex) charset=utf8
// regenconn : DB의 connection이 아무동작을 안하면 DB서버에서는 지정된 시간이후 connection을 강제로 닫습니다.
//             이것을 방지하기위해 사용됩니다.
//             3분마다 connection을 재생성합니다.
func (d *DB) Setup(dbdriver string, conncmd string, regenconn bool) error {
	var err error
	d.Db, err = sql.Open(dbdriver, conncmd)
	if err != nil {
		return err
	} else {
		d.Db.SetMaxOpenConns(5)            // 최대 동작 connection
		d.Db.SetMaxIdleConns(5)            // 최대 생성가능한 기본상태 connection
		d.Db.SetConnMaxLifetime(time.Hour) // 최대 연결시간

		if regenconn {
			d.regenstate = make(chan bool, 1)
			go func() {
				for {
					time.Sleep(time.Minute * 3)
					err := d.regenconn()
					if err != nil {
						fmt.Printf("오류!, 자세히 : %v", err)
					}
				}
			}()
		}
		return nil
	}
}

func (d *DB) SetConnection() error {
	var err error
	d.Conn, err = d.Db.Conn(dbctx)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) Query(query string, args ...interface{}) interface{} {
	var rows *sql.Rows
	var err error
	if !isnil(args) {
		//prepare example  사용후 종료해줘야함! smts.Close()
		//smts,_ := d.Conn.PrepareContext(dbctx, query)
		//rows, err = smts.QueryContext(dbctx, args...)

		// query example
		rows, err = d.Conn.QueryContext(dbctx, query, args...)
		// 위의 두개다 정상작동

	} else {
		rows, err = d.Conn.QueryContext(dbctx, query)
	}
	if err != nil {
		return errors.Errorf("Query err : %v", err)
	} else {
		data, err := d.result(rows)
		if err != nil {
			return err
		} else {
			return fmt.Sprintf("%v", data)
		}
	}
}

func (d DB) result(rows *sql.Rows) (string, error) {
	cols, err := rows.Columns() // get rows col
	if err != nil {
		return "", errors.Errorf("rows col err : %v", err)
	}
	rawResult := make([][]byte, len(cols))
	result := make([]string, len(cols))
	resultdata := []string{}
	dest := make([]interface{}, len(cols))
	for i := range rawResult {
		dest[i] = &rawResult[i] // 결괏값을 가져오기위해 rawresult의 포인터를 dest array에 담음
	}

	for rows.Next() {
		err = rows.Scan(dest...) // 결과값을 읽어서 dest에 대입 (내부코드보면 dest에 정보를 대입함.)
		if err != nil {
			log.Println("Failed to scan row", err)
			return "", nil
		}

		for i, raw := range rawResult {
			if raw == nil {
				result[i] = ""
			} else {
				result[i] = string(raw)
			}
		}
		resultdata = append(resultdata, strings.Join(result, "|")) // 여러줄일경우 고려
	}
	rows.Close()

	return strings.Join(resultdata, "@"), nil // 여러줄일경우 @로 줄구분
}

func (d *DB) regenconn() (err error) {
	defer func() {
		s := recover()
		if s != nil {
			err = errors.Errorf("regen panic error detail : %v", err)
		}
	}()
	var errs error
	var workstate bool
	select {
	case workstate = <-d.regenstate:
	default:
	}
	if workstate {
		d.Conn, errs = d.Db.Conn(dbctx)
		if errs != nil {
			return errs
		}
	}
	return nil

}
func (d *DB) Setregenstate(state bool) {
	d.regenstate <- state
}
