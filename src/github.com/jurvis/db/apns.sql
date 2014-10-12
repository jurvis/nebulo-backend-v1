CREATE TABLE devicetokens (
	id serial not null primary key,
	uuid text not null unique,
	devicetype text not null
);
