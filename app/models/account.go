package models

import "github.com/CodeSyncr/nimbus/database"

// Group is a client group.
type Group struct {
	database.Model
	Name string

	Accounts []Account
}

// Client is a client entity.
type Client struct {
	database.Model
	Name string

	Accounts []Account
}

// Account demonstrates belongsTo relationships, matching the AdonisJS example:
//
//	export default class Account extends Base {
//	    @column()
//	    public groupId: number
//
//	    @column()
//	    public clientId: number
//
//	    @belongsTo(() => Group)
//	    public group: BelongsTo<typeof Group>
//
//	    @belongsTo(() => Client)
//	    public client: BelongsTo<typeof Client>
//	}
type Account struct {
	database.Model
	GroupID         uint
	ClientID        uint
	AccountName     string
	Amount          float64
	TransactionType string
	IsVisible       bool

	Group  *Group
	Client *Client
}
