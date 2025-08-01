# doc-builder

`doc-builder` is a framework for interpolating files into documents.

The primary motivation was to allow source code samples in markdown documents
to have automated testing.

## Example

The following is the config file that created this `README.md` document:

```json
{
  "variables": {
    "project_name": "doc-builder"
  },
  "targets": [
    {
      "source": "README.tmpl.md",
      "output": "README.md",
      "inputs": [
        "assets/intro.md",
        "dbc.json"
      ]
    }
  ]
}
```
