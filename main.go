package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql" // mysql driver
)

type DataMap struct {
	ColumnName    string
	ColumnComment string
	DataType      string
	MaxLength     sql.NullString
	IsNullable    string
	ColumnDefault sql.NullString
	ColumnType    string
}

type Table struct {
	Name string
}

//数据库配置
const (
	userName = "root"
	password = "123456"
	ip       = "127.0.0.1"
	port     = "3306"
	dbName   = "Work"
)

//Db数据库连接池
var DB *sql.DB

func main() {
	outputFile, outputError := os.OpenFile(dbName+".md", os.O_WRONLY|os.O_CREATE, 0666)
	if outputError != nil {
		fmt.Printf("An error occurred with file opening or creation\n")
		return
	}
	defer outputFile.Close()

	outputWriter := bufio.NewWriter(outputFile)

	InitDB()
	defer DB.Close()

	tbs := []Table{}

	quertTable := "SELECT TABLE_NAME Name from information_schema.tables where TABLE_SCHEMA= ?"
	rows, err := DB.Query(quertTable, dbName)
	defer rows.Close()
	if err != nil {
		fmt.Println("DB.Query error:", err)
		return
	}

	for rows.Next() {
		tb := Table{}
		rows.Scan(&tb.Name)
		tbs = append(tbs, tb)
	}

	sql := `SELECT COLUMN_NAME ColumnName, 
	COLUMN_COMMENT ColumnComment, 
	DATA_TYPE DataType, 
	CHARACTER_MAXIMUM_LENGTH MaxLength, 
	IS_NULLABLE IsNullable, 
	COLUMN_DEFAULT ColumnDefault, 
	COLUMN_TYPE ColumnType
	FROM information_schema.columns WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?`

	for _, v := range tbs {
		cols, err := DB.Query(sql, dbName, v.Name)
		defer rows.Close()
		if err != nil {
			fmt.Println("DB.Query error:", err)
			return
		}

		outputWriter.WriteString("#### " + v.Name + "\n")
		outputWriter.WriteString("\n")
		outputWriter.WriteString("| 字段名称    | 字段描述          | 字段类型  | 长度 | 允许空 | 缺省值 |\n")
		outputWriter.WriteString("|-------------|-------------------|-----------|------|--------|--------|\n")

		for cols.Next() {
			m := DataMap{}
			cols.Scan(
				&m.ColumnName,
				&m.ColumnComment,
				&m.DataType,
				&m.MaxLength,
				&m.IsNullable,
				&m.ColumnDefault,
				&m.ColumnType,
			)

			if m.MaxLength.String == "" {
				if m.DataType == "timestamp" {
					m.MaxLength.String = "0"
				} else {
					lenStr := getLen(m.ColumnType)
					if lenStr == "" {
						m.MaxLength.String = "0"
					} else {
						m.MaxLength.String = lenStr
					}

				}
			}

			if m.DataType == "timestamp" {
				m.MaxLength.String = "0"
			}

			if m.MaxLength.String == "" {
				m.MaxLength.String = getLen(m.ColumnType)
			}

			switch m.IsNullable {
			case "":
				m.IsNullable = "是"
			case "NO":
				m.IsNullable = "否"
			}

			outputWriter.WriteString("|" + m.ColumnName + "|" + m.ColumnComment + "|" + m.DataType + "|" + m.MaxLength.String + "|" + m.IsNullable + "|" + m.ColumnDefault.String + "|\n")
		}
		outputWriter.WriteString("\n")
		outputWriter.WriteString("\n")
		cols.Close()
	}

	outputWriter.Flush()
}

func InitDB() {
	//构建连接："用户名:密码@tcp(IP:端口)/数据库?charset=utf8"
	path := strings.Join([]string{userName, ":", password, "@tcp(", ip, ":", port, ")/", dbName, "?charset=utf8"}, "")

	DB, err := sql.Open("mysql", path)
	if err != nil {
		panic(err)
	}

	//验证连接
	if err = DB.Ping(); err != nil {
		fmt.Println("opon database fail")
		return
	}
	fmt.Println("connnect success")
}

func getLen(s string) string {
	if len(s) == 0 {
		return "0"
	}

	rec := make([]byte, 0, 0)
	for i := 0; i < len(s); i++ {
		if s[i] == '(' {
			for j := i + 1; j < len(s); j++ {
				if s[j] == ')' {
					goto End
				}
				rec = append(rec, s[j])
			}
		}
	}
End:
	return string(rec)
}
