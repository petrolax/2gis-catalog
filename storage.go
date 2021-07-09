package main

import (
	"sort"
	"sync"

	"github.com/jmoiron/sqlx"
)

type Company struct {
	Name    string `json:"name" db:"name"`
	Phones  string `json:"phones" db:"phones"`
	Address string `db:"address"`
	Rubric  int64  `db:"rubric"`
}

// Данный тип создан для того чтобы сортировать массив организаций
type Companies []Company

func (c Companies) Len() int           { return len(c) }
func (c Companies) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c Companies) Less(i, j int) bool { return c[i].Name < c[j].Name }

type JSONCompany struct {
	Name    string  `json:"name" db:"name"`
	Phones  string  `json:"phones" db:"phones"`
	Address string  `db:"address"`
	Rubrics []int64 `json:"rubrics"`
}

type Building struct {
	Address     string        `json:"address"`
	Coordinates string        `json:"coordinates"`
	Companies   []JSONCompany `json:"companies"`
}

type Storage interface {
	InsertBuilding(b *Building) error
	GetCompany(company_id int64) ([]Company, error)
	GetCompaniesFromBuilding(building_id int64) ([]Company, error)
	GetCompaniesFromRubric(rubric_id int64) ([]Company, error)
}

type BuildingStorage struct {
	db *sqlx.DB
	sync.Mutex
}

func NewBuildingStorage(db *sqlx.DB) *BuildingStorage {
	return &BuildingStorage{
		db: db,
	}
}

func (bs *BuildingStorage) InsertBuilding(b *Building) error {
	requestAddBuilding := `INSERT INTO handbook.building (address, coordinates) VALUES ($1, $2)`
	requestAddCompany := `INSERT INTO handbook.company (name, phones, building_id) VALUES ($1, $2, $3)`
	requestAddRubricOfCompany := `INSERT INTO handbook.rubricsofcompany (company_id, rubric_id) VALUES ($1, $2)`
	bs.db.MustExec(requestAddBuilding, b.Address, b.Coordinates)
	var buildingId int64
	err := bs.db.Get(&buildingId, `SELECT building_id FROM handbook.building WHERE address = $1`, b.Address)
	if err != nil {
		return err
	}

	for _, comp := range b.Companies {
		bs.db.MustExec(requestAddCompany, comp.Name, comp.Phones, buildingId)
		var companyID int64
		err := bs.db.Get(&companyID, `SELECT company_id FROM handbook.company WHERE name = $1`, comp.Name)
		if err != nil {
			return err
		}
		for _, rubric := range comp.Rubrics {
			//Описать, почему не проверяешь наличие рубрики в таблице
			bs.db.MustExec(requestAddRubricOfCompany, companyID, rubric)
		}
	}
	return nil
}

func (bs BuildingStorage) GetCompany(companyId int64) ([]Company, error) {
	var comps []Company
	request := `SELECT hc.name, hc.phones, 
		hb.address as address, 
		hr.rubric_id as rubric
	FROM handbook.company as hc
	INNER JOIN handbook.rubricsofcompany as hr ON hr.company_id = hc.company_id
	INNER JOIN handbook.building as hb ON hb.building_id=hc.building_id
	WHERE hc.company_id= $1`
	err := bs.db.Select(&comps, request, companyId)
	if err != nil {
		return nil, err
	}
	return comps, nil
}

func (bs BuildingStorage) GetCompaniesFromBuilding(buildingId int64) ([]Company, error) {
	var comps []Company
	request := `SELECT hc.name, hc.phones, 
		hb.address as address, 
		hr.rubric_id as rubric
	FROM handbook.company as hc
	INNER JOIN handbook.rubricsofcompany as hr ON hr.company_id = hc.company_id
	INNER JOIN handbook.building as hb ON hb.building_id=hc.building_id
	WHERE hc.building_id= $1`
	err := bs.db.Select(&comps, request, buildingId)
	if err != nil {
		return nil, err
	}
	return comps, nil
}

func (bs BuildingStorage) GetCompaniesFromRubric(rubricId int64) ([]Company, error) {
	var comps []Company
	request := `SELECT hc.name, hc.phones, 
		hb.address as address, 
		hroc.rubric_id as rubric
	FROM handbook.company as hc
	INNER JOIN handbook.rubricsofcompany as hroc ON hroc.company_id = hc.company_id
	INNER JOIN handbook.building as hb ON hb.building_id=hc.building_id
	WHERE hroc.rubric_id= $1`
	err := bs.db.Select(&comps, request, rubricId)
	if err != nil {
		return nil, err
	}
	request = `SELECT hc.name, hc.phones, 
		hb.address as address, 
		hroc.rubric_id as rubric
	FROM handbook.company as hc
	INNER JOIN handbook.rubricsofcompany as hroc ON hc.company_id = hroc.company_id
	INNER JOIN handbook.building as hb ON hb.building_id=hc.building_id
	inner join handbook.rubric as hr on hr.rubric_id = hroc.rubric_id 
	WHERE hr.parent_id= $1`
	err = bs.db.Select(&comps, request, rubricId)
	if err != nil {
		return nil, err
	}
	sort.Sort(Companies(comps))
	return comps, nil
}