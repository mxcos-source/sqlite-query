package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 数据库文件名
	dbFile := "jfxtData.db"
	
	// 检查数据库文件是否存在
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		log.Fatalf("错误: 数据库文件 '%s' 不存在于当前目录", dbFile)
	}

	fmt.Printf("正在打开数据库: %s\n", dbFile)
	
	// 打开数据库连接
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 测试数据库连接
	err = db.Ping()
	if err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	// 执行你的 SQL 查询
	query := `
SELECT
	u.id,
	u.type,
	s.name,
	(SELECT sum(m.top_up) FROM memberCardReFillRecord m WHERE m.member_id = u.id) AS money
FROM
	userMember u 
	LEFT JOIN user s ON u.user_id = s.id
WHERE
	u.type = 5`

	fmt.Println("执行查询...")
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("SQL执行失败: %v", err)
	}
	defer rows.Close()

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		log.Fatalf("获取列信息失败: %v", err)
	}

	// 创建表格输出器
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	
	// 打印表头
	fmt.Println("\n查询结果:")
	for _, col := range columns {
		fmt.Fprintf(w, "%s\t", col)
	}
	fmt.Fprintln(w)

	// 打印分隔线
	for range columns {
		fmt.Fprintf(w, "------------\t")
	}
	fmt.Fprintln(w)

	// 准备接收数据的切片
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	rowCount := 0
	// 遍历结果集
	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			log.Printf("读取数据失败: %v", err)
			continue
		}

		// 打印每一行数据
		for i := range columns {
			var value interface{}
			val := values[i]
			
			// 处理 NULL 值
			if val == nil {
				value = "NULL"
			} else {
				// 处理字节数组（字符串）
				b, ok := val.([]byte)
				if ok {
					value = string(b)
				} else {
					value = val
				}
			}
			fmt.Fprintf(w, "%v\t", value)
		}
		fmt.Fprintln(w)
		rowCount++
	}

	w.Flush()
	fmt.Printf("\n总共查询到 %d 行数据\n", rowCount)

	if err = rows.Err(); err != nil {
		log.Fatalf("遍历结果集失败: %v", err)
	}
}
