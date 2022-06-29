# How to create a template
For the moment you have to create a `scripts` section. Then under the `scripts`
you have to create `onmount` or `oncreate` section. Then specify the OS that the
script is going to be executed. If you want to execute the script in every OS use universal instead.

`owl_config.toml` config example
```toml
[scripts]

[scripts.oncreate]
[scripts.oncreate.universal]
init_git = ["git init"]

[scripts.onmount]
[scripts.onmount.universal]
init_package = [
    "go mod init {{ .ModuleName }}",
    "go mod tidy"
]
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
                "init_package": ["go mod init {{ .ModuleName }}", "go mod tidy"]
            }
        }
    }
}
```
