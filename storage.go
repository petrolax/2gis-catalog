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

func (c Companies) Len() int { 
	return len(c)
}

func (c Companies) Swap(i, j int) {
	c[i], c[j] = c[j], c[i] 
}

// Реализация функции не совпадает с её наименованием, но данное наименование необходимо для функции sort.Sort()
func (c Companies) Less(i, j int) bool {
	if c[i].Name != c[j].Name {
		return c[i].Name < c[j].Name 
	} else {
		if c[i].Address != c[j].Address {
			return c[i].Address < c[j].Address
		} else {
			return false
		}
	}
}

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
	GetCompany(companyId int64) ([]Company, error)
	GetCompaniesFromBuilding(buildingId int64) ([]Company, error)
	GetCompaniesFromRubric(rubricId int64) ([]Company, error)
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
	bs.Lock()
	defer bs.Unlock()

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
	bs.Lock()
	defer bs.Unlock()

	var comps []Company
	request := `SELECT hc.name, hc.phones, 
		hb.address AS address, 
		hr.rubric_id AS rubric
	FROM handbook.company AS hc
	INNER JOIN handbook.rubricsofcompany AS hr ON hr.company_id = hc.company_id
	INNER JOIN handbook.building AS hb ON hb.building_id=hc.building_id
	WHERE hc.company_id= $1`
	err := bs.db.Select(&comps, request, companyId)
	if err != nil {
		return nil, err
	}
	return comps, nil
}

func (bs BuildingStorage) GetCompaniesFromBuilding(buildingId int64) ([]Company, error) {
	bs.Lock()
	defer bs.Unlock()

	var comps []Company
	request := `SELECT hc.name, hc.phones, 
		hb.address AS address, 
		hr.rubric_id AS rubric
	FROM handbook.company AS hc
	INNER JOIN handbook.rubricsofcompany AS hr ON hr.company_id = hc.company_id
	INNER JOIN handbook.building AS hb ON hb.building_id=hc.building_id
	WHERE hc.building_id= $1`
	err := bs.db.Select(&comps, request, buildingId)
	if err != nil {
		return nil, err
	}
	return comps, nil
}

func (bs BuildingStorage) GetCompaniesFromRubric(rubricId int64) ([]Company, error) {
	bs.Lock()
	defer bs.Unlock()

	var comps []Company
	request := `SELECT hc.name, hc.phones, 
		hb.address AS address, 
		hroc.rubric_id AS rubric
	FROM handbook.company AS hc
	INNER JOIN handbook.rubricsofcompany AS hroc ON hroc.company_id = hc.company_id
	INNER JOIN handbook.building AS hb ON hb.building_id=hc.building_id
	WHERE hroc.rubric_id= $1`
	if err := bs.db.Select(&comps, request, rubricId); err != nil {
		return nil, err
	}
	request = `SELECT hc.name, hc.phones, 
		hb.address AS address, 
		hroc.rubric_id AS rubric
	FROM handbook.company AS hc
	INNER JOIN handbook.rubricsofcompany AS hroc ON hc.company_id = hroc.company_id
	INNER JOIN handbook.building AS hb ON hb.building_id=hc.building_id
	INNER JOIN handbook.rubric AS hr ON hr.rubric_id = hroc.rubric_id 
	WHERE hr.parent_id= $1`
	if err := bs.db.Select(&comps, request, rubricId); err != nil {
		return nil, err
	}

	requestID := `SELECT hr.rubric_id
	FROM handbook.company AS hc
	INNER JOIN handbook.rubricsofcompany AS hroc ON hc.company_id = hroc.company_id
	INNER JOIN handbook.building AS hb ON hb.building_id=hc.building_id
	INNER JOIN handbook.rubric AS hr ON hr.rubric_id = hroc.rubric_id 
	GROUP BY hr.rubric_id
	HAVING hr.parent_id = $1`
	var rubricIDs []int64
	if err := bs.db.Select(&rubricIDs, requestID, rubricId); err != nil {
		return nil, err
	}

	compsLen := len(comps)
	// Цикл необходим чтобы считать все рубрики каждой фирмы
	for {
		for _, rubric := range rubricIDs {
			if err := bs.db.Select(&comps, request, rubric); err != nil {
				return nil, err
			}
		}
		if compsLen == len(comps) {
			break
		}

		var nextRubricIDs []int64
		for _, rubric := range rubricIDs {
			if err := bs.db.Select(&nextRubricIDs, requestID, rubric); err != nil {
				return nil, err
			}
		}
		if len(nextRubricIDs) == 0 { 
			break
		}
		rubricIDs = nextRubricIDs
	}

	sort.Sort(Companies(comps))
	return comps, nil
}
