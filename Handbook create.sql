create schema handbook;

create table handbook.building (
	building_id serial primary key,
	address text NOT NULL,
	coordinates text NOT NULL
);

create table handbook.rubric (
	rubric_id serial primary key,
	parent_id integer references handbook.rubric(rubric_id),
	name text NOT NULL
);

create table handbook.company (
	company_id serial primary key,
	name text NOT NULL,
	phones text NOT NULL,
	building_id serial references handbook.building(building_id)	
);

create table handbook.rubricsofcompany (
	company_id serial references handbook.company(company_id),
	rubric_id serial references handbook.rubric(rubric_id)
);