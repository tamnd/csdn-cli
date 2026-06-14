package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/csdn-cli/csdn"
)

// sortMap translates user-facing flag values to CSDN API sort integers.
var sortMap = map[string]int{
	"relevance": 0,
	"newest":    1,
	"hot":       2,
}

func (a *App) searchCmd() *cobra.Command {
	var (
		sort    string
		page    int
		ctype   string
	)
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search CSDN blog posts",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n := a.effectiveLimit(20)
			q := args[0]

			sortInt, ok := sortMap[sort]
			if !ok {
				sortInt = 0
			}

			a.progressf("searching CSDN for %q (sort=%s, type=%s, page=%d)...", q, sort, ctype, page)

			opts := csdn.SearchOpts{
				Query: q,
				Type:  ctype,
				Sort:  sortInt,
				Page:  page,
			}
			articles, total, err := a.client.Search(cmd.Context(), opts)
			if err != nil {
				return mapFetchErr(err)
			}

			if len(articles) > n {
				articles = articles[:n]
			}

			a.progressf("found %d total results, showing %d", total, len(articles))
			return a.renderOrEmpty(articles, len(articles))
		},
	}
	cmd.Flags().StringVar(&sort, "sort", "relevance", "sort order: relevance|newest|hot")
	cmd.Flags().IntVar(&page, "page", 1, "page number (1-based)")
	cmd.Flags().StringVar(&ctype, "type", "blog", "content type: blog|ask|download")
	return cmd
}
