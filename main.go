package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/tabwriter"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 定义命令行参数
	dbFile := flag.String("db", "", "SQLite数据库文件路径（默认当前目录查找.db文件）")
	query := flag.String("sql", "", "要执行的SQL查询语句")
	flag.Parse()

	// 如果没有指定数据库文件，查找当前目录的.db文件
	if *dbFile == "" {
		files, err := filepath.Glob("*.db")
		if err != nil {
			log.Fatal("查找数据库文件错误:", err)
		}
		if len(files) == 0 {
			log.Fatal("未找到.db文件，请使用 -db 参数指定数据库文件")
		}
		*dbFile = files[0]
		fmt.Printf("使用数据库文件: %s\n", *dbFile)
	}

	// 验证SQL语句
	if *query == "" {
		flag.Usage()
		log.Fatal("必须使用 -sql 参数指定查询语句")
	}

	// 打开数据库连接
	db, err := sql.Open("sqlite3", *dbFile)
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	// 执行查询
	rows, err := db.Query(*query)
	if err != nil {
		log.Fatal("SQL执行失败:", err)
	}
	defer rows.Close()

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		log.Fatal("获取列信息失败:", err)
	}

	// 创建表格输出器
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	
	// 打印表头
	for _, col := range columns {
		fmt.Fprintf(w, "%s\t", col)
	}
	fmt.Fprintln(w)

	// 打印分隔线
	for range columns {
		fmt.Fprintf(w, "--------\t")
	}
	fmt.Fprintln(w)

	// 准备接收数据的切片
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// 遍历结果集
	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			log.Fatal("读取数据失败:", err)
		}

		// 打印每一行数据
		for i := range columns {
			var value interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				value = string(b)
			} else {
				value = val
			}
			fmt.Fprintf(w, "%v\t", value)
		}
		fmt.Fprintln(w)
	}

	w.Flush()

	if err = rows.Err(); err != nil {
		log.Fatal("遍历结果集失败:", err)
	}
}
