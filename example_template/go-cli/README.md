# How to create a template
For the moment you have to create a `scripts` section. Then under the `scripts`
you have to create `onmount` or `oncreate` section. Then specify the OS that the
script is going to be executed. If you want to execute the script in every OS use universal instead.

## oncreate
oncreate scripts are executed when the template root folder is created but before the files are copied.

## onmount
onmount scripts are executed when the template was copied.

## configuration examples
`owl_config.toml` config example
```toml
[scripts]

[scripts.oncreate]
[scripts.oncreate.universal]
init_git = ["git init"]

[scripts.onmount]
[scripts.onmount.universal]
init_package = ["go mod init {{ .ModuleName }}"]
```

`owl_config.json` config example
```json
{
    "scripts": {
        "onmount": {
            "universal": {
                "init_git": ["git init"]
            }
        },
        "oncreate": {
            "universal": {
                "init_package": ["go mod init {{ .ModuleName }}"]
            }
        }
    }
}
```

## Variables
When you create template, you can pass a map filled with variables that can be used by the scripts.
