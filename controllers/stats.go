package controllers

import m "github.com/chiefsend/api/models"

func numberOfShares() (int, error) {
	// getDatabase
	db, err := m.GetDatabase()
	if err != nil {
		return -1, err
	}

	var shares []m.Share
	if err := db.Find(&shares).Error; err != nil {
		return -1, err
	}

	return 5, nil
}

func heatMap(sh m.Share) error {
	return nil
}
