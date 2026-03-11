/*
|--------------------------------------------------------------------------
| Hashing Configuration
|--------------------------------------------------------------------------
|
| Algorithm and cost for password hashing.
|
*/

package config

var Hash HashConfig

type HashConfig struct {
	Driver     string
	BcryptCost int
}

func loadHash() {
	Hash = HashConfig{
		Driver:     env("HASH_DRIVER", "bcrypt"),
		BcryptCost: envInt("HASH_BCRYPT_COST", 10),
	}
}
