# Go development

Before any Go coding, review, debugging, troubleshooting, or setup task, load the `samber/cc-skills-golang@golang-how-to` skill first — it routes to whichever other Go skills the task needs.

Read the ownership and snapshot contracts in `doc.go` before changing range storage. Stored backing arrays are immutable even though `Set` has mutating methods.

Use the independent point model in the tests to verify range algorithms. Keep the module and its test suite dependent only on the Go standard library.
