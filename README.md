# coar-vocabularies-website

These notes are in development...

1. Sources for the COAR Vocabularies website
2. Sources and binaries for the tool used to generate the website

## Notes on development
Source skos files must be in `ntriples` format
Source SKOS files *must* contain *one* `conceptScheme`

Recommended to run the SKOS files through Skosify, to clean up the syntax and optionally add some semantics. Also, exports from iQVoc lack a conceptScheme, and Skosify can automatically add one of these.

```bash
skosify \
  --namespace="http://purl.org/coar/access_right/" \
  --label="Access Rights" \
  --set-modified \
  --mark-top-concepts \
  --narrower \
  --eliminate-redundancy \
  --cleanup-classes \
  --cleanup-properties \
  --cleanup-unreachable \
  -o ./skosified/access_rights.nt \
  ./access_rights.nt
```

```bash
skosify \
  --namespace="http://purl.org/coar/resource_type/" \
  --label="Resource Types" \
  --set-modified \
  --mark-top-concepts \
  --narrower \
  --eliminate-redundancy \
  --cleanup-classes \
  --cleanup-properties \
  --cleanup-unreachable \
  -o ./skosified/resource_types.nt \
  ./resource_types.nt
```

```bash
skosify \
  --namespace="http://purl.org/coar/version/" \
  --label="Version Types" \
  --set-modified \
  --mark-top-concepts \
  --narrower \
  --eliminate-redundancy \
  --cleanup-classes \
  --cleanup-properties \
  --cleanup-unreachable \
  -o ./skosified/version_types.nt \
  ./version_types.nt
```

