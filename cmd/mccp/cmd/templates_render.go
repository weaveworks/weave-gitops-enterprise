package cmd

import (
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/wks/cmd/mccp/pkg/adapters"
	"github.com/weaveworks/wks/cmd/mccp/pkg/formatter"
	"github.com/weaveworks/wks/cmd/mccp/pkg/templates"
)

var templatesRenderCmd = &cobra.Command{
	Use:     "render",
	Short:   "Render CAPI template",
	Example: `mccp templates render <template-name>`,
	RunE:    templatesRenderCmdRun,
	Args:    cobra.ExactArgs(1),
}

var (
	listTemplateParameters bool
	parameterValues        []string
)

func init() {
	templatesCmd.AddCommand(templatesRenderCmd)
	templatesRenderCmd.PersistentFlags().BoolVar(&listTemplateParameters, "list-parameters", false, "The CAPI templates HTTP API endpoint")
	templatesRenderCmd.PersistentFlags().StringArrayVar(&parameterValues, "set", []string{}, "Set parameter values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
}

func templatesRenderCmdRun(cmd *cobra.Command, args []string) error {
	r, err := adapters.NewHttpTemplateRetriever(endpoint, resty.New())
	if err != nil {
		return err
	}

	if listTemplateParameters {
		w := formatter.NewTableWriter()
		defer w.Flush()
		return templates.ListTemplateParameters(args[0], r, w)
	}
	vals := make(map[string]string)
	for _, v := range parameterValues {
		kv := strings.SplitN(v, "=", 2)
		if len(kv) == 2 {
			vals[kv[0]] = kv[1]
		}
	}
	return templates.RenderTemplate(args[0], vals, r, os.Stdout)
}
