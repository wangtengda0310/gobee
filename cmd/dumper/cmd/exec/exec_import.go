package exec

import (
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/wangtengda0310/gobee/lvan/cmd/dumper/cmd/internal"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/write"
	"github.com/wangtengda0310/gobee/lvan/pkg/template"
)

var importExecCmd = &cobra.Command{
	Use:   "import",
	Short: "导入数据",
	Long:  `connect mysql import`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.Println("PersistentPreRun import")
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		log.Println("PersistentPreRunE import")
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("import called")

		c := dbParams(cmd)
		database := c.Database
		table := c.Table
		log.Println("config", c)

		VisitImport(func(mysqlCli *exec.Cmd) {

			for _, input := range args {

				// 检查input是否为空
				if input == "" {
					log.Panic("import-dir is required")
				}

				var records []dump.Record
				records = internal.TransImport(input)
				for _, record := range records {

					for _, beforeImport := range dump.BeforeImportEachRecordsCallback {
						beforeImport(record)
					}
					strRecord := map[string]string{}
					for k, v := range record {
						strRecord[k] = string(v)
					}

					// 解析template
					sql, err := template.RenderTemplate(input, strRecord)
					if err != nil {
						log.Panic(err)
					}
					temp, err := os.CreateTemp("", "lvan.import.sql.")
					if err != nil {
						log.Panic(err)
					}
					log.Println(temp.Name())
					{
						// 存储新sql到文件
						err = os.WriteFile(temp.Name(), []byte(sql), 0755)
						if err != nil {
							log.Panic(err)
						}

						// 调用MySQL << file
						mysqlCli.Stdin = temp
						v, err := mysqlCli.CombinedOutput()
						if err != nil {
							anies := string(v)
							log.Panic(anies, err)
						}
						log.Println(string(v))
					}

					err = temp.Close()
					if err != nil {
						log.Panic(err)
					}
					err = os.Remove(temp.Name())
					if err != nil {
						log.Panic(err)
					}
				}

				write.Console(records)

				log.Println(database, table)

			}

		}, "", "")
	},
}

func init() {
	execCmd.AddCommand(importExecCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	//importCmd.Flags().BoolP("help", "?", false, "Help message for import")

}
