package vcardgo

import (
	"fmt"
	"reflect"
	"strings"
)

// Struct used for schema definition. See [StringSchemaV4] as an example.
//
// vCard schema may differ slightly between applications e.g. Android vs Outlook.
//
// Note that in .vcf specification each record may have different schema depending
// on it's version, which means you can provide multiple schemas to Decoder.
//
// Implementation field will be used by reflect to extract/inject field names and types
// from/into used-defined types.
//
// For example:
//
//		  // Schema-type defined by this library
//		  type StringSchemaV4 struct {
//		      FN string `vCard:"required"`
//			     ...
//		  }
//
//		  // User-defined type has to use the name and the type from the Schema
//	      // or tag a field with `vCard:"FN"`
//		  type MyUser struct {
//		      FN string
//		  }
//
//		  vCard.Marshal(MyUser{FN: "Nick"})
type Schema struct {
	version        string       // Schema version (e.g. "2.1", "3.0", "4.0")
	implementation reflect.Type // Type used by reflect to compare with used-defined types.
}

// Creates a new schema for generic definition
func NewSchema[T any](version string) Schema {
	return Schema{version: version, implementation: reflect.TypeFor[T]()}
}

// Creates a new schema from provided instance of implementation struct
func NewSchemaFrom(version string, implementation any) Schema {
	return Schema{version: version, implementation: reflect.TypeOf(implementation)}
}

// Simple Schema for string-based vCard 4.0
var DefaultSchemaV4 = NewSchema[StringSchemaV4]("4.0")

// Simple Schema for string-based vCard 3.0
var DefaultSchemaV3 = NewSchema[StringSchemaV3]("3.0")

// Simple Schema for string-based vCard 2.1
var DefaultSchemaV2_1 = NewSchema[StringSchemaV2_1]("2.1")

// Caches schema parameters for fast access in [Marshal]/[Unmarshal]
func (s Schema) Prepare() PreparedSchema {
	kind := s.implementation.Kind()
	if kind != reflect.Struct {
		panic(fmt.Sprintf("vCard: schema implementation %s is not a struct", kind))
	}

	m := make(map[string]SchemaField)

	requiredSlice := []string{}

	for i := range s.implementation.NumField() {
		field := s.implementation.Field(i)

		name := field.Name
		typ := field.Type
		optional := true

		if strings.Contains(field.Tag.Get("vCard"), "required") {
			optional = false
			requiredSlice = append(requiredSlice, name)
		}

		m[name] = SchemaField{typ: typ, optional: optional}
	}

	return PreparedSchema{
		version:        s.version,
		implementation: s.implementation,
		hmap:           m,
		requiredFields: requiredSlice,
	}
}

type PreparedSchema struct {
	version        string                 // User-defined version from Schema type.
	implementation reflect.Type           // original type used for reflect.
	hmap           map[string]SchemaField // all fields in easy-to-use format.
	requiredFields []string               // cached for better performance.
}

type SchemaField struct {
	typ      reflect.Type
	optional bool
}


// Simple vCard v4.0 schema implementation from https://en.wikipedia.org/wiki/VCard
//
// Note that this struct can be used as argument in vCard.Unmarshal without
// the need to provide user-defined type.
//
// The difference between v4.0 and v3.0 is that N property is optional.
type StringSchemaV4 struct {
	ADR         string // A structured representation of the delivery address for the person.
	AGENT       string // Information about another person who will act on behalf of this one.
	ANNIVERSARY string // Defines the person's anniversary.
	BDAY        string // Date of birth of the individual.

	// BEGIN:VCARD - All vCards must start with this property.

	CALADRURI    string   // A URL to use for sending a scheduling request to the person's calendar.
	CALURI       string   // A URL to the person's calendar.
	CATEGORIES   string // A list of "tags" that can be used to describe the person.
	CLASS        string   // Describes the sensitivity of the information in the vCard.
	CLIENTPIDMAP string   // Used for synchronizing different revisions of the same vCard.
	EMAIL        string   // The address for electronic mail communication

	// END:VCARD - All vCards must end with this property.

	FBURL  string // Defines a URL that shows when the person is "free" or "busy" on their calendar.
	FN     string `vCard:"required"` // The formatted name string.
	GENDER string // Defines the person's gender.
	GEO    string // Specifies a latitude and longitude.
	IMPP   string // Defines an instant messenger handle.
	KEY    string // The public encryption key associated with the person.
	KIND   string // Defines the type of entity that this vCard represents.

	// Represents the actual text that should be put on the mailing label. Not supported in version 4.0.
	// Instead, this information is stored in the LABEL parameter of the ADR property.
	LABEL string

	LANG   string // Defines a language that the person speaks.
	LOGO   string // An image or graphic of the logo of the organization that is associated with the individual.
	MAILER string // Type of email program used.
	MEMBER string // Defines a member that is part of the group that this vCard represents.
	N      string // A structured representation of the name of the person

	// Provides a textual representation of the SOURCE property. Not to be confused with N property
	// which defines person's name.
	NAME string

	NICKNAME string // One or more descriptive/familiar names.
	NOTE     string // Comment that is associated with the person.
	ORG      string // The name and optionally the unit(s) of the organization associated with the person.

	PHOTO   string // An image of the individual. It may point to an external URL or may be embedded as a Base64.
	PRODID  string // The identifier for the product that created the vCard object.
	PROFILE string // States that the vCard is a vCard.

	RELATED string // Another entity that the person is related to.
	REV     string // A timestamp for the last time the vCard was updated.
	ROLE    string // The role, occupation, or business category of the person within an organization.

	// Defines a string that should be used when an application sorts this vCard in some way.
	// Not supported in 4.0. Instead, this is stored in the SORT-AS parameter of the N and/or ORG properties.
	SORT_STRING string

	SOUND  string // Specifies the pronunciation of the FN property.
	SOURCE string // A URL that can be used to get the latest version of this vCard.
	TEL    string // The canonical number string for a telephone number.

	TITLE string // Specifies the job title, functional position or function of the individual.
	TZ    string // The time zone of the person.
	UID   string // Specifies a persistent, globally unique identifier associated with the person.
	URL   string // A URL pointing to a website that represents the person in some way.

	// The version of the vCard specification.
	// Ignored by [Marshal].
	VERSION string

	XML string // Any XML data that is attached to the vCard.
}

// Simple vCard v3.0 schema implementation from https://en.wikipedia.org/wiki/VCard
//
// Note that this struct can be used as argument in vCard.Unmarshal without
// the need to provide user-defined type.
//
// The difference between v4.0 and v3.0 is that N property is required to be set.
type StringSchemaV3 struct {
	ADR         string // A structured representation of the delivery address for the person.
	AGENT       string // Information about another person who will act on behalf of this one.
	ANNIVERSARY string // Defines the person's anniversary.
	BDAY        string // Date of birth of the individual.

	// BEGIN:VCARD - All vCards must start with this property.

	CALADRURI    string   // A URL to use for sending a scheduling request to the person's calendar.
	CALURI       string   // A URL to the person's calendar.
	CATEGORIES   []string // A list of "tags" that can be used to describe the person.
	CLASS        string   // Describes the sensitivity of the information in the vCard.
	CLIENTPIDMAP string   // Used for synchronizing different revisions of the same vCard.
	EMAIL        string   // The address for electronic mail communication

	// END:VCARD - All vCards must end with this property.

	FBURL  string // Defines a URL that shows when the person is "free" or "busy" on their calendar.
	FN     string `vCard:"required"` // The formatted name string.
	GENDER string // Defines the person's gender.
	GEO    string // Specifies a latitude and longitude.
	IMPP   string // Defines an instant messenger handle.
	KEY    string // The public encryption key associated with the person.
	KIND   string // Defines the type of entity that this vCard represents.

	// Represents the actual text that should be put on the mailing label. Not supported in version 4.0.
	// Instead, this information is stored in the LABEL parameter of the ADR property.
	LABEL string

	LANG   string // Defines a language that the person speaks.
	LOGO   string // An image or graphic of the logo of the organization that is associated with the individual.
	MAILER string // Type of email program used.
	MEMBER string // Defines a member that is part of the group that this vCard represents.
	N      string `vCard:"required"` // A structured representation of the name of the person

	// Provides a textual representation of the SOURCE property. Not to be confused with N property
	// which defines person's name.
	NAME string

	NICKNAME string // One or more descriptive/familiar names.
	NOTE     string // Comment that is associated with the person.
	ORG      string // The name and optionally the unit(s) of the organization associated with the person.

	PHOTO   string // An image of the individual. It may point to an external URL or may be embedded as a Base64.
	PRODID  string // The identifier for the product that created the vCard object.
	PROFILE string // States that the vCard is a vCard.

	RELATED string // Another entity that the person is related to.
	REV     string // A timestamp for the last time the vCard was updated.
	ROLE    string // The role, occupation, or business category of the person within an organization.

	// Defines a string that should be used when an application sorts this vCard in some way.
	// Not supported in 4.0. Instead, this is stored in the SORT-AS parameter of the N and/or ORG properties.
	SORT_STRING string

	SOUND  string // Specifies the pronunciation of the FN property.
	SOURCE string // A URL that can be used to get the latest version of this vCard.
	TEL    string // The canonical number string for a telephone number.

	TITLE string // Specifies the job title, functional position or function of the individual.
	TZ    string // The time zone of the person.
	UID   string // Specifies a persistent, globally unique identifier associated with the person.
	URL   string // A URL pointing to a website that represents the person in some way.

	// VERSION - The version of the vCard specification.

	XML string // Any XML data that is attached to the vCard.
}

// Simple vCard v2.1 schema implementation from https://en.wikipedia.org/wiki/VCard
//
// Note that this struct can be used as argument in vCard.Unmarshal without
// the need to provide user-defined type.
//
// The difference between v2.1 and v3.0 is that FN property is optional.
type StringSchemaV2_1 struct {
	ADR         string // A structured representation of the delivery address for the person.
	AGENT       string // Information about another person who will act on behalf of this one.
	ANNIVERSARY string // Defines the person's anniversary.
	BDAY        string // Date of birth of the individual.

	// BEGIN:VCARD - All vCards must start with this property.

	CALADRURI    string   // A URL to use for sending a scheduling request to the person's calendar.
	CALURI       string   // A URL to the person's calendar.
	CATEGORIES   []string // A list of "tags" that can be used to describe the person.
	CLASS        string   // Describes the sensitivity of the information in the vCard.
	CLIENTPIDMAP string   // Used for synchronizing different revisions of the same vCard.
	EMAIL        string   // The address for electronic mail communication

	// END:VCARD - All vCards must end with this property.

	FBURL  string // Defines a URL that shows when the person is "free" or "busy" on their calendar.
	FN     string // The formatted name string.
	GENDER string // Defines the person's gender.
	GEO    string // Specifies a latitude and longitude.
	IMPP   string // Defines an instant messenger handle.
	KEY    string // The public encryption key associated with the person.
	KIND   string // Defines the type of entity that this vCard represents.

	// Represents the actual text that should be put on the mailing label. Not supported in version 4.0.
	// Instead, this information is stored in the LABEL parameter of the ADR property.
	LABEL string

	LANG   string // Defines a language that the person speaks.
	LOGO   string // An image or graphic of the logo of the organization that is associated with the individual.
	MAILER string // Type of email program used.
	MEMBER string // Defines a member that is part of the group that this vCard represents.
	N      string `vCard:"required"` // A structured representation of the name of the person

	// Provides a textual representation of the SOURCE property. Not to be confused with N property
	// which defines person's name.
	NAME string

	NICKNAME string // One or more descriptive/familiar names.
	NOTE     string // Comment that is associated with the person.
	ORG      string // The name and optionally the unit(s) of the organization associated with the person.

	PHOTO   string // An image of the individual. It may point to an external URL or may be embedded as a Base64.
	PRODID  string // The identifier for the product that created the vCard object.
	PROFILE string // States that the vCard is a vCard.

	RELATED string // Another entity that the person is related to.
	REV     string // A timestamp for the last time the vCard was updated.
	ROLE    string // The role, occupation, or business category of the person within an organization.

	// Defines a string that should be used when an application sorts this vCard in some way.
	// Not supported in 4.0. Instead, this is stored in the SORT-AS parameter of the N and/or ORG properties.
	SORT_STRING string

	SOUND  string // Specifies the pronunciation of the FN property.
	SOURCE string // A URL that can be used to get the latest version of this vCard.
	TEL    string // The canonical number string for a telephone number.

	TITLE string // Specifies the job title, functional position or function of the individual.
	TZ    string // The time zone of the person.
	UID   string // Specifies a persistent, globally unique identifier associated with the person.
	URL   string // A URL pointing to a website that represents the person in some way.

	// VERSION - The version of the vCard specification.

	XML string // Any XML data that is attached to the vCard.
}
