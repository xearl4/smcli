package wallet

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/spacemeshos/smcli/common"
	"github.com/tyler-smith/go-bip39"
)

// Wallet is the basic data structure.
type Wallet struct {
	//keystore string
	//password string
	//unlocked bool
	Meta    walletMetadata `json:"meta"`
	Secrets walletSecrets  `json:"crypto"`

	// this is not persisted
	//masterKeypair *EDKeyPair
}

// EncryptedWalletFile is the encrypted representation of the wallet on the filesystem
type EncryptedWalletFile struct {
	Meta    walletMetadata         `json:"meta"`
	Secrets walletSecretsEncrypted `json:"crypto"`
}

type walletMetadata struct {
	DisplayName string `json:"displayName"`
	Created     string `json:"created"`
	GenesisID   string `json:"genesisID"`
	//NetID       int    `json:"netId"`

	// is this needed?
	//Type WalletType
	//RemoteAPI string
}

type hexEncodedCiphertext []byte

func (c *hexEncodedCiphertext) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(*c))
}

func (c *hexEncodedCiphertext) UnmarshalJSON(data []byte) (err error) {
	var hexString string
	if err = json.Unmarshal(data, &hexString); err != nil {
		return
	}
	*c, err = hex.DecodeString(hexString)
	return
}

type walletSecretsEncrypted struct {
	Cipher       string               `json:"cipher"`
	CipherText   hexEncodedCiphertext `json:"cipherText"`
	CipherParams struct {
		IV hexEncodedCiphertext `json:"iv"`
	} `json:"cipherParams"`
	KDF       string `json:"kdf"`
	KDFParams struct {
		DKLen      int                  `json:"dklen"`
		Hash       string               `json:"hash"`
		Salt       hexEncodedCiphertext `json:"salt"`
		Iterations int                  `json:"iterations"`
	} `json:"kdfparams"`
}

type walletSecrets struct {
	Mnemonic string       `json:"mnemonic"`
	Accounts []*EDKeyPair `json:"accounts"`
}

func NewMultiWalletRandomMnemonic(n int) (*Wallet, error) {
	// generate a new, random mnemonic
	e, err := bip39.NewEntropy(ed25519.SeedSize * 8)
	if err != nil {
		return nil, err
	}
	m, err := bip39.NewMnemonic(e)
	if err != nil {
		return nil, err
	}

	return NewMultiWalletFromMnemonic(m, n)
}

func NewMultiWalletFromMnemonic(m string, n int) (*Wallet, error) {
	if n < 0 || n > common.MaxAccountsPerWallet {
		return nil, fmt.Errorf("invalid number of accounts")
	}
	if !bip39.IsMnemonicValid(m) {
		return nil, fmt.Errorf("invalid mnemonic")
	}
	// TODO: add option for user to provide passphrase
	// https://github.com/spacemeshos/smcli/issues/18
	seed := bip39.NewSeed(m, "")
	accounts, err := accountsFromSeed(seed, n)
	if err != nil {
		return nil, err
	}
	return walletFromMnemonicAndAccounts(m, accounts)
}

func walletFromMnemonicAndAccounts(m string, kp []*EDKeyPair) (*Wallet, error) {
	displayName := "Main Wallet"
	createTime := common.NowTimeString()

	w := &Wallet{
		Meta: walletMetadata{
			DisplayName: displayName,
			Created:     createTime,
			// TODO: set correctly
			GenesisID: "",
		},
		Secrets: walletSecrets{
			Mnemonic: m,
			Accounts: kp,
		},
	}
	return w, nil
}

// accountsFromSeed generates one or more accounts from a given seed. Accounts use sequential HD paths.
func accountsFromSeed(seed []byte, n int) (accounts []*EDKeyPair, err error) {
	masterKeyPair, err := NewMasterKeyPair(seed[:ed25519.SeedSize])
	if err != nil {
		return
	}
	for i := 0; i < n; i++ {
		acct, err := masterKeyPair.NewChildKeyPair(i)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acct)
	}
	return
}

func (w *Wallet) Mnemonic() string {
	return w.Secrets.Mnemonic
}
