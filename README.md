# doc-builder

This is an introduction to the doc-builder.

## Example

The following is the config file that created this `README.md` document:

```json
{
  "env": {
    "project_name": "doc-builder"
  },
  "targets": [
    {
      "source": "README.tmpl.md",
      "output": "README.md",
      "inputs": [
        {
          "source": "assets/intro.md",
          "output": "#env"
        },
        {
          "source": "dbc.json",
          "output": "#env"
        }
      ]
    }
  ]
}
```
