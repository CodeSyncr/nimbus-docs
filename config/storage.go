/*
|--------------------------------------------------------------------------
| Storage / Drive Configuration
|--------------------------------------------------------------------------
|
| File storage driver for uploads and generated files.
|
*/

package config

var Storage StorageConfig

type StorageConfig struct {
	Driver string
	Local  LocalStorageConfig
}

type LocalStorageConfig struct {
	Root string
}

func loadStorage() {
	Storage = StorageConfig{
		Driver: env("STORAGE_DRIVER", "local"),
		Local: LocalStorageConfig{
			Root: env("STORAGE_ROOT", "storage"),
		},
	}
}
