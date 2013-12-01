package main

import (
	"fmt"
	"github.com/aybabtme/color/brush"
	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
	"github.com/crosbymichael/pulls"
	"os"
)

var (
	m *pulls.Maintainer
)

func displayAllPullRequests(c *cli.Context, state string) {
	// FIXME: Pass a filter to the Getpullrequests method
	prs, err := m.GetPullRequests(state)
	prs, err = m.FilterPullRequests(prs, c)
	if err != nil {
		writeError("Error getting pull requests %s", err)
	}
	fmt.Printf("%c[2K\r", 27)
	pulls.DisplayPullRequests(c, prs, c.Bool("no-trunc"))
}

func addComment(number, comment string) {
	cmt, err := m.AddComment(number, comment)
	if err != nil {
		writeError("%s", err)
	}
	pulls.DisplayCommentAdded(cmt)
}

func repositoryInfoCmd(c *cli.Context) {
	r, err := m.Repository()
	if err != nil {
		writeError("%s", err)
	}
	fmt.Fprintf(os.Stdout, "Name: %s\nForks: %d\nStars: %d\nIssues: %d\n", r.Name, r.Forks, r.Watchers, r.OpenIssues)
}

func mergeCmd(c *cli.Context) {
	number := c.Args()[0]
	merge, err := m.MergePullRequest(number, c.String("m"))
	if err != nil {
		writeError("%s", err)
	}
	if merge.Merged {
		fmt.Fprintf(os.Stdout, "%s\n", brush.Green(merge.Message))
	} else {
		writeError("%s", err)
	}
}

func checkoutCmd(c *cli.Context) {
	number := c.Args()[0]
	pr, _, err := m.GetPullRequest(number, false)
	if err != nil {
		writeError("%s", err)
	}
	if err := m.Checkout(pr); err != nil {
		writeError("%s", err)
	}
}

// Approve a pr by adding a LGTM to the comments
func approveCmd(c *cli.Context) {
	number := c.Args().First()
	if _, err := m.AddComment(number, "LGTM"); err != nil {
		writeError("%s", err)
	}
	fmt.Fprintf(os.Stdout, "Pull request %s approved\n", brush.Green(number))
}

// This is the top level command for
// working with prs
func mainCmd(c *cli.Context) {
	if !c.Args().Present() {
		state := "open"
		if c.Bool("closed") {
			state = "closed"
		}
		displayAllPullRequests(c, state)
		return
	}

	var (
		number  = c.Args().Get(0)
		comment = c.String("comment")
	)

	if comment != "" {
		addComment(number, comment)
		return
	}
	pr, comments, err := m.GetPullRequest(number, true)
	if err != nil {
		writeError("%s", err)
	}
	pulls.DisplayPullRequest(pr, comments)
}

func main() {
	app := cli.NewApp()

	app.Name = "pulls"
	app.Usage = "Manage github pull requests for project maintainers"
	app.Version = "0.0.1"

	client := gh.NewClient()

	config := loadConfig()
	if config.Token != "" {
		client.WithToken(config.Token)
	}

	org, name, err := getOriginUrl()
	if err != nil {
		writeError("%s", err)
	}
	t, err := pulls.NewMaintainer(client, org, name)
	if err != nil {
		writeError("%s", err)
	}
	m = t

	loadCommands(app)

	app.Run(os.Args)
}
