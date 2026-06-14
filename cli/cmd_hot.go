package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) hotCmd() *cobra.Command {
	var page int
	cmd := &cobra.Command{
		Use:   "hot",
		Short: "Show hot articles on CSDN",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(20)
			a.progressf("fetching CSDN hot articles (page %d, size %d)...", page, n)

			articles, err := a.client.Hot(cmd.Context(), page, n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(articles, len(articles))
		},
	}
	cmd.Flags().IntVar(&page, "page", 0, "page number (0-based)")
	return cmd
}
