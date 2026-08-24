module dummy-fs-provider

go 1.25.0

require github.com/Tarafagat/asterion-plugin-contract v0.0.0-00010101000000-000000000000

// Vive dentro del propio repo de asterion-plugin-contract, así que la ruta
// relativa hacia arriba es fija — un plugin de verdad, en su propio repo,
// usaría "../asterion-plugin-contract" (ver sdk/go/scaffold/templates/go.mod.tmpl).
replace github.com/Tarafagat/asterion-plugin-contract => ../..
