package models

import "github.com/CodeSyncr/nimbus/database"

// User demonstrates hasMany, hasOne, and manyToMany relationships.
//
// AdonisJS equivalent:
//
//	export default class User extends BaseModel {
//	    @hasMany(() => Post)
//	    public posts: HasMany<typeof Post>
//
//	    @hasOne(() => Profile)
//	    public profile: HasOne<typeof Profile>
//
//	    @manyToMany(() => Team)
//	    public teams: ManyToMany<typeof Team>
//	}
type User struct {
	database.Model
	Name  string
	Email string

	Posts   []Post
	Profile *Profile
	Teams   []Team
}

// Post demonstrates belongsTo and hasMany.
//
// AdonisJS equivalent:
//
//	export default class Post extends BaseModel {
//	    @column()
//	    public userId: number
//
//	    @belongsTo(() => User)
//	    public user: BelongsTo<typeof User>
//
//	    @hasMany(() => Comment)
//	    public comments: HasMany<typeof Comment>
//	}
type Post struct {
	database.Model
	UserID    uint
	Title     string
	Body      string
	Published bool

	User     *User
	Comments []Comment
}

// Comment demonstrates belongsTo.
type Comment struct {
	database.Model
	PostID uint
	UserID uint
	Body   string

	Post *Post
	User *User
}

// Profile demonstrates belongsTo (inverse of hasOne).
type Profile struct {
	database.Model
	UserID uint
	Bio    string
	Avatar string

	User *User
}

// Team demonstrates manyToMany (inverse side).
// Pivot table: team_users (auto-generated, alphabetical order).
type Team struct {
	database.Model
	Name string

	Users []User
}
