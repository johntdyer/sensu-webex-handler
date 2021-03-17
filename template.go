package main

const inccidentTemplate = `{
  {{.MessageTarget}}
  "markdown": "<blockquote class='{{.MessageColor}}'> {{.Emoji}} {{.MessageStatus}} <br/>
      <b>Check Name:</b> {{.CheckName}}
      {{if (ne .MessageStatus "Resolved") }}
      &nbsp;&nbsp;&nbsp;&nbsp; <b>Execution Time:</b> {{.CheckExecutionTime}} <br/>
      {{end}}
      <b>Entity:</b> {{.EntityName}}      <br/>
      {{if (ne .MessageStatus "Resolved") }}
      <b>Check output:</b> {{.CheckOutput}} <br/>
      <b>History:</b> {{.History}} <br/>
    {{end}}</blockquote>",
  "attachments": [
    {
      "contentType": "application/vnd.microsoft.card.adaptive",
      "content": {
        "type": "AdaptiveCard",
        "version": "1.0",
        "body": [
          {
            "type": "Container",
            "items": [{
              "type": "ColumnSet",
              "columns": [{
                "type": "Column",
                "width": "100px",
                "items": [{
                  "type": "TextBlock",
                  "text": "{{.Emoji}} {{.MessageStatus}}",
                  "size": "Medium",
                  "isSubtle": true
                }]
                },
                {
                  "type": "Column",
                  "width": "300px",
                  "items": [{
                    "type": "TextBlock",
                    "text": "**Check Name**: [{{.CheckName}}](foo)"
                  }]
                }
              ],
              "horizontalAlignment": "Left"
            }
            ],
            "spacing": "Medium",
            "horizontalAlignment": "Left",
            "style": "default"
          },
          {
            "type": "ColumnSet",
            "columns": [{
                "type": "Column",
                "width": "5px",
                "items": [{
                  "type": "Image",
                  "altText": "",
                  "url": "{{.BucketName}}/{{.MessageColor}}.png",
                  "spacing": "Medium",
                  "backgroundColor": "green"
                }],
                "spacing": "None",
                "horizontalAlignment": "Center",
                "backgroundImage": {
                  "url": "{{.BucketName}}/{{.MessageColor}}.png",
                  "fillMode": "RepeatVertically",
                  "horizontalAlignment": "Center"
                }
              }

              ,{
                "type": "Column",
                "width": "stretch",
                "items": [{
                  "type": "ColumnSet",
                  "columns": [{
                    "type": "Column",
                    "width": "stretch",
                    "items": [{
                      "type": "FactSet",
                      "facts": [
                        {
                          "title": "**Entity:** ",
                          "value": "[{{.EntityName}}](foo)"
                        }
                        {{if (ne .MessageStatus "Resolved") }}
                        ,
                        {
                          "title": "Time",
                          "value": "{{.CheckExecutionTime}}"
                        },
                        {
                          "title": "History",
                          "value": "{{.History}}"
                        }
                        {{end}}
                      ]
                    }]
                  }]
                }]
              }

            ]
          }
          {{if (ne .MessageStatus "Resolved") }}
          ,
          {
            "type": "Container",
            "items": [{
              "type": "Container",
              "items": [{
                "type": "ColumnSet",
                "columns": [{
                  "type": "Column",
                  "width": "stretch",
                  "items": [{
                    "type": "TextBlock",
                    "text": "**Check Output**: {{.CheckOutput}}",
                    "wrap": true,
                    "color": "Attention",
                    "separator": true,
                    "horizontalAlignment": "Left",
                    "size": "Small"
                  }]
                }]
              }]
            }]
          }
          {{end}}
        ]
      }
    }
  ]
}`
