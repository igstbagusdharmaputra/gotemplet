package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v3"
)

var rootCmd = &cobra.Command{
	Use:     "tmpl",
	Short:   "Root command for tmpl interaction",
	Long:    `Root command for tmpl interaction.`,
	Version: "1.0.0",
}


var generateCmd = &cobra.Command{
	Use: "generate",
	Short: "Using generate template",
	Long: "Using generate template",
	RunE: func(cmd *cobra.Command, args []string) error {
		command := cmd.Flags()
		templatePath, _ := command.GetString("templatePath")
		subTemplatePath, _ := command.GetString("subTemplatePath")
		bindingdataPath,_ := command.GetString("dataPath")
		outputPath,_ := command.GetString("outputPath")
		
		templatePath, _, _ = getTemplatePath(templatePath,subTemplatePath)	
		// fmt.Println(path)
		// fmt.Println(tmpPath)
		// fmt.Println(repo)
		listTemplates, err := listTemplatePath(templatePath)
		if err != nil {
			return err
		}
		var context map[string]interface{}

		if bindingdataPath != "" {
			bindingData, err := parseBindingData(bindingdataPath)
			if err != nil {
				return err
			}
			context = bindingData
		}
		for _, p := range listTemplates {
			
			err := renderAndSave(p,templatePath,outputPath,context)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

func getTemplatePath(templatePath string, subTemplatePath string) (string, string, bool){

	if _, ok := StringStartWith(templatePath, []string{"http://", "https://", "file://", "git://", "ssh://"}, true); ok {
		return fmt.Sprintf("%s/%s",templatePath, subTemplatePath), "", true
	}else{
		return path.Join(templatePath, subTemplatePath), "", false
		
	}
}

func listTemplatePath(templatePath string) ([]string, error) {
	templatePaths := make([]string,0)
	pathInfo, err := os.Stat(templatePath)
	if err != nil {
		return nil, err
	}
	if pathInfo.IsDir() {
		
		dataPathInfo, err := os.ReadDir(templatePath)
		if err != nil {
			return nil, err
		}
		
		for _, data := range dataPathInfo {
			
			dataPaths, err := listTemplatePath(path.Join(templatePath, data.Name()))
			
			if err != nil {
				return nil, err
			}
			
			templatePaths = append(templatePaths, dataPaths...)
			
		}
	}else{
		templatePaths = append(templatePaths, templatePath)
		
	}
	
	return templatePaths,nil
}
func renderAndSave(templateFilePath,templatePath,outputFilePath string, bindingDatas map[string]interface{}) error {
	
	templateFilePath = strings.Replace(templateFilePath, filepath.Dir(templatePath), outputFilePath, 1)
	// t := fasttemplate.New(templateFilePath, "{{ ", " }}")
	// resultTemplateFilePath, err := mustache.Render(templateFilePath, bindingDatas...)
	// resultTemplateFilePath := t.ExecuteString(bindingDatas)
	var tpl bytes.Buffer
	t := template.Must(template.New("path").Parse(templateFilePath))
	if err := t.Execute(&tpl, bindingDatas); err != nil {
		return err
	}
	resultTemplateFilePath := tpl.String()
	// if err != nil {
	// 	return err
	// }
	
	dirPath := filepath.Dir(resultTemplateFilePath)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		//be careful, must be 07xx
		err = os.MkdirAll(dirPath, 0744)
		if err != nil {
			return err
		}
	}
	err := os.WriteFile(resultTemplateFilePath, []byte("hi"), 0644)
	if err != nil {
		return err
	}
	fmt.Println(resultTemplateFilePath)
	return nil
}
func StringStartWith(str string, subStrs []string, caseSensitive bool) (string, bool) {
	str = strings.Trim(str, " \n\t")
	if !caseSensitive {
		str = strings.ToLower(str)
	}
	for _, sub := range subStrs {
		prefix := sub
		if !caseSensitive {
			prefix = strings.ToLower(prefix)
		}
		if strings.HasPrefix(str, prefix) {
			return sub, true
		}
	}

	return "", false
}

func parseBindingData(bindingDataPath string) (map[string]interface{}, error) {
	b, err := os.ReadFile(bindingDataPath)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	if strings.HasSuffix(strings.ToLower(bindingDataPath), "yaml") {
		if err := yaml.Unmarshal(b, &data); err != nil {
			return nil, err
		}
	} else if strings.HasSuffix(strings.ToLower(bindingDataPath), "json") {
		if err := json.Unmarshal(b, &data); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("not support binding data format")
	}
	return data, nil
}



func init(){
	generateCmd.Flags().StringP("templatePath", "t", "", "Directory name or repo template")
	generateCmd.Flags().StringP("subTemplatePath", "s", "", "sub template path")
	generateCmd.Flags().StringP("dataPath", "d","","Path file data binding")
	generateCmd.Flags().StringP("outputPath","o","","Output generate template")
	rootCmd.AddCommand(generateCmd)

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}