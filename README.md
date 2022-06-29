# cluster-controller

```yaml
---
# путь до каталога, где должны храниться манифесты
manifests_path: "/etc/kubernetes/manifests"
# спаисок манифестов, которые необходимы на хост машине
manifests:
  # имя манифеста, манифест будет сохранен в файл  /etc/kubernetes/manifests/{name}.yaml
  - name: "etcd"
  # путь до файла с шаблном для данного манифеста
  - template_path: "/path/to/manifests/template"
  # аргументы которые будут добавлены в шаблон манифеста
  - args:
      - "arg1"
      - "arg2"
```
