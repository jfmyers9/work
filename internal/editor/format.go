package editor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jfmyers9/work/internal/model"
)

func MarshalIssue(issue model.Issue) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Title: %s\n", issue.Title)
	fmt.Fprintf(&b, "Type: %s\n", issue.Type)
	fmt.Fprintf(&b, "Priority: %d\n", issue.Priority)
	fmt.Fprintf(&b, "Labels: %s\n", strings.Join(issue.Labels, ", "))
	fmt.Fprintf(&b, "Assignee: %s\n", issue.Assignee)

	b.WriteString("\n")
	if issue.Description != "" {
		b.WriteString(issue.Description)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	fmt.Fprintf(&b, "# ID: %s | Status: %s | Created: %s\n",
		issue.ID, issue.Status, issue.Created.Format("2006-01-02"))
	b.WriteString("# Lines starting with '#' are ignored.\n")
	b.WriteString("# Leave the description section empty to clear it.\n")

	return b.String()
}

func UnmarshalIssue(text string) (title, description, issueType, assignee string, priority int, labels []string, err error) {
	var headerLines []string
	var bodyLines []string
	pastHeader := false

	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}
		if !pastHeader {
			if line == "" {
				pastHeader = true
				continue
			}
			headerLines = append(headerLines, line)
		} else {
			bodyLines = append(bodyLines, line)
		}
	}

	headers := make(map[string]string)
	for _, line := range headerLines {
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		headers[strings.TrimSpace(key)] = strings.TrimSpace(val)
	}

	title = headers["Title"]
	if title == "" {
		err = fmt.Errorf("title is required")
		return
	}

	issueType = headers["Type"]
	assignee = headers["Assignee"]

	if p, ok := headers["Priority"]; ok {
		priority, _ = strconv.Atoi(p)
	}

	if raw, ok := headers["Labels"]; ok && raw != "" {
		for _, l := range strings.Split(raw, ",") {
			l = strings.TrimSpace(l)
			if l != "" {
				labels = append(labels, l)
			}
		}
	}

	description = strings.TrimSpace(strings.Join(bodyLines, "\n"))
	return
}
