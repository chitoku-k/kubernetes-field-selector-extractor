Kubernetes Field Selector Extractor
===================================

This application extracts all the available [Field Selectors][] from Kubernetes.
Since the list of implemented field selectors is neither exposed via any APIs
nor declared anywhere, this application statically finds every call for
[AddFieldLabelConversionFunc][] like shown below and looks for hard-coded labels
by retrieving ASTs.

```go
err = scheme.AddFieldLabelConversionFunc(SchemeGroupVersion.WithKind("Node"),
    func(label, value string) (string, string, error) {
        switch label {
        case "metadata.name":
            return label, value, nil
        case "spec.unschedulable":
            return label, value, nil
        default:
            return "", "", fmt.Errorf("field label not supported: %s", label)
        }
    },
)
```

Cited from: [pkgs/apis/core/v1/conversion.go (kubernetes/kubernetes)](https://github.com/kubernetes/kubernetes/blob/v1.21.2/pkg/apis/core/v1/conversion.go#L60-L71)

## Set up

```bash
$ git clone https://github.com/chitoku-k/kubernetes-field-selector-extractor
$ cd kubernetes-field-selector-extractor
$ git submodule update --init
$ go build -o kubernetes-field-selector-extractor
```

## Usage

```
$ ./kubernetes-field-selector-extractor ./kubernetes/pkg/apis > selectors.json
```

## Known Limitations

Unsupported style of coding such as the following cases emit warnings.

* Variable kinds are not supported (such as `for` loop)
  * [pkg/apis/batch/v1/beta1/conversion.go (kubernetes/kubernetes)](https://github.com/kubernetes/kubernetes/blob/v1.21.2/pkg/apis/batch/v1beta1/conversion.go#L28-L42)
* Variable labels are not supported (such as referencing `map`)
  * [pkg/apis/events/v1/conversion.go (kubernetes/kubernetes)](https://github.com/kubernetes/kubernetes/blob/v1.21.2/pkg/apis/events/v1/conversion.go#L63-L87)
  * [pkg/apis/events/v1beta1/conversion.go (kubernetes/kubernetes)](https://github.com/kubernetes/kubernetes/blob/v1.21.2/pkg/apis/events/v1beta1/conversion.go#L63-L87)

## License

Kubernetes is licensed under the Apache License, Version 2.0.

[Field Selectors]:             https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/
[AddFieldLabelConversionFunc]: https://pkg.go.dev/k8s.io/apimachinery/pkg/runtime#Scheme.AddFieldLabelConversionFunc
