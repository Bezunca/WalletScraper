package cei

import (
	"WalletScraper/internal/config"
	"WalletScraper/internal/database"
	"WalletScraper/internal/models"
	"WalletScraper/internal/utils"
	ceiModels "github.com/Bezunca/ceilib/models"
	"github.com/Bezunca/ceilib/scraper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

func ScrapeDividends(ceiCredentials ceiModels.CEI, userID primitive.ObjectID) (*[]mongo.WriteModel, error) {
	log.Printf("Gettting Dividends for %s", ceiCredentials.User)

	dividend, err := scraper.GetUserDividends(ceiCredentials)
	if err != nil{
		return nil, err
	}

	allDividends := append(dividend.Credited, dividend.Provisioned...)
	allDividendsMongo := make([]models.MongoDividends, 0, len(dividend.Provisioned) + len(dividend.Credited))
	for dividend := range allDividends {
		allDividendsMongo = append(allDividendsMongo, models.MongoDividends{
			DividendStats: allDividends[dividend],
			UserID: userID,
		})
	}

	if len(allDividendsMongo) > 0 {
		dividendModels := make([]mongo.WriteModel, 0, len(allDividendsMongo))
		for idx := range allDividendsMongo {
			currDividend := allDividendsMongo[idx]
			if currDividend.Date == nil{
				log.Printf("Ignoring Dividend Without Date: %s", currDividend)
				continue
			}

			mongoDividend, err := utils.ToDoc(currDividend)
			if err != nil {
				return nil, err
			}

			dividendModels = append(
				dividendModels,
				mongo.NewUpdateOneModel().SetFilter(
					bson.D{
						{Key: "data.symbol", Value: currDividend.Symbol},
						{Key: "data.date", Value: currDividend.Date},
						{Key: "data.type", Value: currDividend.Type},
						{Key: "data.base_quantity", Value: currDividend.BaseQuantity},
						{Key: "data.price_factor", Value: currDividend.PriceFactor},
						{Key: "data.gross_income", Value: currDividend.GrossIncome},
						{Key: "data.net_income", Value: currDividend.NetIncome},
						{Key: "user_id", Value: currDividend.UserID},
					},
				).SetUpdate(mongoDividend).SetUpsert(true),
			)
		}

		return &dividendModels, nil
	}

	return &[]mongo.WriteModel{}, nil
}

func ScrapeTrades(mongoClient *mongo.Client, ceiCredentials ceiModels.CEI, userID primitive.ObjectID) ([]interface{}, error) {
	configs := config.Get()
	log.Printf("Gettting Trades for %s", ceiCredentials.User)
	tradeCollection := mongoClient.Database(configs.ApplicationDatabase).Collection("user_trades")

	trades, err := scraper.GetUserTrades(ceiCredentials)
	if err != nil{
		return nil, err
	}

	lastUpdate, err := database.GetLastUpdateTime(*tradeCollection, bson.D{{Key: "user_id", Value: userID}})
	if err != nil{
		return nil, err
	}

	tradesMongo := make([]interface{}, 0, len(trades))
	for idx := range trades {
		trade := trades[idx]
		if lastUpdate == nil || lastUpdate.Before(trade.Date){
			docTradeMongo, err := utils.ToDoc(models.MongoTrades{
				Trade: trades[idx],
				UserID: userID,
			})
			if err != nil{
				return nil, err
			}
			tradesMongo = append(tradesMongo, docTradeMongo.Map())
		}
	}

	return tradesMongo, nil
}

func ScrapePortfolio(ceiCredentials ceiModels.CEI, userID primitive.ObjectID) (*[]mongo.WriteModel, error) {
	log.Printf("Gettting Portfolio for %s", ceiCredentials.User)

	portfolio, err := scraper.GetUserPortfolio(ceiCredentials)
	if err != nil{
		return nil, err
	}

	portfolioMongo := make([]models.MongoPortfolio, 0, len(portfolio))
	for idx := range portfolio {
		portfolioMongo = append(portfolioMongo, models.MongoPortfolio{
			Asset: portfolio[idx],
			UserID: userID,
		})
	}

	if len(portfolioMongo) > 0 {
		portfolioModels := make([]mongo.WriteModel, 0, len(portfolioMongo))
		for idx := range portfolioMongo {
			currPortfolio := portfolioMongo[idx]

			mongoPortfolio, err := utils.ToDoc(currPortfolio)
			if err != nil {
				return nil, err
			}

			portfolioModels = append(
				portfolioModels,
				mongo.NewUpdateOneModel().SetFilter(
					bson.D{
						{Key: "data.symbol", Value: currPortfolio.Symbol},
						{Key: "user_id", Value: currPortfolio.UserID},
					},
				).SetUpdate(mongoPortfolio).SetUpsert(true),
			)
		}

		return &portfolioModels, nil
	}

	return &[]mongo.WriteModel{}, nil
}