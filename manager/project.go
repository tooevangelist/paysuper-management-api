package manager

import (
	"errors"
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-management-api/database/dao"
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/sidmal/slug"
	"go.uber.org/zap"
	"strings"
	"time"
)

const (
	minPaymentAmount float64 = 0
	maxPaymentAmount float64 = 15000

	projectErrorMerchantNotHaveProjects = "merchant not have projects"
	projectErrorAccessDeniedToProject   = "one or more projects in filter are not a projects of merchant"
	projectErrorNotFound                = "project with specified identifier not found"
	projectErrorNotHasPaymentMethods    = "project not has connected payment methods"
)

type ProjectManager struct {
	*Manager

	merchantManager       *MerchantManager
	currencyManager       *CurrencyManager
	paymentMethodsManager *PaymentMethodManager
}

func InitProjectManager(database dao.Database, logger *zap.SugaredLogger) *ProjectManager {
	pm := &ProjectManager{
		Manager:               &Manager{Database: database, Logger: logger},
		merchantManager:       InitMerchantManager(database, logger),
		currencyManager:       InitCurrencyManager(database, logger),
		paymentMethodsManager: InitPaymentMethodManager(database, logger),
	}

	return pm
}

func (pm *ProjectManager) Create(ps *model.ProjectScalar) (*model.Project, error) {
	p := &model.Project{
		Id:                         bson.NewObjectId(),
		Merchant:                   ps.Merchant,
		Name:                       ps.Name,
		CallbackProtocol:           ps.CallbackProtocol,
		CreateInvoiceAllowedUrls:   ps.CreateInvoiceAllowedUrls,
		IsAllowDynamicNotifyUrls:   ps.IsAllowDynamicNotifyUrls,
		IsAllowDynamicRedirectUrls: ps.IsAllowDynamicRedirectUrls,
		OnlyFixedAmounts:           ps.OnlyFixedAmounts,
		FixedPackage:               pm.processFixedPackages(ps.FixedPackage, true),
		SecretKey:                  ps.SecretKey,
		URLCheckAccount:            ps.URLCheckAccount,
		URLProcessPayment:          ps.URLProcessPayment,
		URLRedirectFail:            ps.URLRedirectFail,
		URLRedirectSuccess:         ps.URLRedirectSuccess,
		SendNotifyEmail:            ps.SendNotifyEmail,
		NotifyEmails:               ps.NotifyEmails,
		IsActive:                   ps.IsActive,
		CreatedAt:                  time.Now(),
		UpdatedAt:                  time.Now(),
	}

	p.LimitsCurrency = p.Merchant.Currency
	p.CallbackCurrency = p.Merchant.Currency

	if ps.LimitsCurrency != nil {
		if c := pm.currencyManager.FindByCodeInt(*ps.LimitsCurrency); c != nil {
			p.LimitsCurrency = c
		}
	}

	if ps.CallbackCurrency != nil {
		if c := pm.currencyManager.FindByCodeInt(*ps.CallbackCurrency); c != nil {
			p.CallbackCurrency = c
		}
	}

	p.MinPaymentAmount = minPaymentAmount
	p.MaxPaymentAmount = maxPaymentAmount

	if ps.MinPaymentAmount != nil {
		p.MinPaymentAmount = *ps.MinPaymentAmount
	}

	if ps.MaxPaymentAmount != nil {
		p.MaxPaymentAmount = *ps.MaxPaymentAmount
	}

	err := pm.Database.Repository(TableProject).InsertProject(p)

	if err != nil {
		pm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableProject, err)
	}

	return p, err
}

func (pm *ProjectManager) Update(p *model.Project, pn *model.ProjectScalar) (*model.Project, error) {
	p.CreateInvoiceAllowedUrls = pn.CreateInvoiceAllowedUrls
	p.NotifyEmails = pn.NotifyEmails
	p.UpdatedAt = time.Now()
	p.FixedPackage = pm.processFixedPackages(pn.FixedPackage, false)

	if p.Name != pn.Name {
		p.Name = pn.Name
	}

	if p.CallbackProtocol != pn.CallbackProtocol {
		p.CallbackProtocol = pn.CallbackProtocol
	}

	if p.IsAllowDynamicNotifyUrls != pn.IsAllowDynamicNotifyUrls {
		p.IsAllowDynamicNotifyUrls = pn.IsAllowDynamicNotifyUrls
	}

	if p.IsAllowDynamicRedirectUrls != pn.IsAllowDynamicRedirectUrls {
		p.IsAllowDynamicRedirectUrls = pn.IsAllowDynamicRedirectUrls
	}

	if p.OnlyFixedAmounts != pn.OnlyFixedAmounts {
		p.OnlyFixedAmounts = pn.OnlyFixedAmounts
	}

	if p.SecretKey != pn.SecretKey {
		p.SecretKey = pn.SecretKey
	}

	if p.URLCheckAccount != pn.URLCheckAccount {
		p.URLCheckAccount = pn.URLCheckAccount
	}

	if p.URLProcessPayment != pn.URLProcessPayment {
		p.URLProcessPayment = pn.URLProcessPayment
	}

	if p.URLRedirectFail != pn.URLRedirectFail {
		p.URLRedirectFail = pn.URLRedirectFail
	}

	if p.URLRedirectSuccess != pn.URLRedirectSuccess {
		p.URLRedirectSuccess = pn.URLRedirectSuccess
	}

	if p.SendNotifyEmail != pn.SendNotifyEmail {
		p.SendNotifyEmail = pn.SendNotifyEmail
	}

	if p.IsActive != pn.IsActive {
		p.IsActive = pn.IsActive
	}

	if pn.LimitsCurrency != nil && (p.LimitsCurrency == nil || p.LimitsCurrency.CodeInt != *pn.LimitsCurrency) {
		if c := pm.currencyManager.FindByCodeInt(*pn.LimitsCurrency); c != nil {
			p.LimitsCurrency = c
		}
	}

	if pn.CallbackCurrency != nil && (p.CallbackCurrency == nil || p.CallbackCurrency.CodeInt != *pn.CallbackCurrency) {
		if c := pm.currencyManager.FindByCodeInt(*pn.CallbackCurrency); c != nil {
			p.CallbackCurrency = c
		}
	}

	if pn.MinPaymentAmount != nil && p.MinPaymentAmount != *pn.MinPaymentAmount {
		p.MinPaymentAmount = *pn.MinPaymentAmount
	}

	if pn.MaxPaymentAmount != nil && p.MaxPaymentAmount != *pn.MaxPaymentAmount {
		p.MaxPaymentAmount = *pn.MaxPaymentAmount
	}

	err := pm.Database.Repository(TableProject).UpdateProject(p)

	if err != nil {
		pm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableProject, err)
	}

	return p, err
}

func (pm *ProjectManager) Delete(p *model.Project) error {
	p.IsActive = false
	p.UpdatedAt = time.Now()

	return pm.Database.Repository(TableProject).UpdateProject(p)
}

func (pm *ProjectManager) FindProjectsByMerchantIdAndName(mId bson.ObjectId, pName string) *model.Project {
	p, err := pm.Database.Repository(TableProject).FindProjectByMerchantIdAndName(mId, pName)

	if err != nil {
		return nil
	}

	return p
}

func (pm *ProjectManager) FindProjectsByMerchantId(mId string, limit int, offset int) []*model.Project {
	p, err := pm.Database.Repository(TableProject).FindProjectsByMerchantId(mId, limit, offset)

	if err != nil {
		pm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableProject, err)
	}

	if p == nil {
		return []*model.Project{}
	}

	return p
}

func (pm *ProjectManager) FindProjectsMainData(mId string) map[string]string {
	p := pm.FindProjectsByMerchantId(mId, model.DefaultLimit, model.DefaultOffset)

	pmd := make(map[string]string)

	if len(p) > 0 {
		for _, v := range p {
			pmd[v.Id.Hex()] = v.Name
		}
	}

	return pmd
}

func (pm *ProjectManager) FindProjectById(id string) *model.Project {
	bId := bson.ObjectIdHex(id)
	p, err := pm.Database.Repository(TableProject).FindProjectById(bId)

	if err != nil {
		pm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableProject, err)
	}

	if len(p.FixedPackage) <= 0 {
		return p
	}

	for _, packages := range p.FixedPackage {
		for _, p := range packages {
			if p.CurrencyInt == 0 {
				continue
			}

			p.Currency = pm.currencyManager.FindByCodeInt(p.CurrencyInt)
		}
	}

	return p
}

func (pm *ProjectManager) processFixedPackages(fixedPackages map[string][]*model.FixedPackage, isNew bool) map[string][]*model.FixedPackage {
	for _, packages := range fixedPackages {
		for _, p := range packages {
			if isNew == true {
				p.CreatedAt = time.Now()
			}

			p.UpdatedAt = time.Now()

			if p.CurrencyInt > 0 {
				if c := pm.currencyManager.FindByCodeInt(p.CurrencyInt); c != nil {
					p.Currency = c
				}
			}

			p.Id = strings.TrimSpace(p.Id)

			if p.Id == "" {
				p.Id = slug.MakeLang(p.Name, slug.DefaultLang, model.FixedPackageSlugSeparator)
			}
		}
	}

	return fixedPackages
}

func (pm *ProjectManager) FilterProjects(mId string, fProjects []bson.ObjectId) (map[bson.ObjectId]string, *model.Merchant, error) {
	mProjects := pm.FindProjectsByMerchantId(mId, model.DefaultLimit, model.DefaultOffset)

	if len(mProjects) <= 0 {
		return nil, nil, errors.New(projectErrorMerchantNotHaveProjects)
	}

	var fp = make(map[bson.ObjectId]string)

	for _, p := range mProjects {
		fp[p.Id] = p.Name
	}

	if len(fProjects) <= 0 {
		return fp, mProjects[0].Merchant, nil
	}

	fp1 := make(map[bson.ObjectId]string)

	for _, p := range fProjects {
		if _, ok := fp[p]; !ok {
			continue
		}

		fp1[p] = fp[p]
	}

	if len(fp1) <= 0 {
		return nil, nil, errors.New(projectErrorAccessDeniedToProject)
	}

	return fp1, mProjects[0].Merchant, nil
}

func (pm *ProjectManager) GetProjectPaymentMethods(projectId bson.ObjectId) ([]*model.PaymentMethod, error) {
	p := pm.FindProjectById(projectId.Hex())

	if p == nil {
		return nil, errors.New(projectErrorNotFound)
	}

	if len(p.PaymentMethods) <= 0 {
		return nil, errors.New(projectErrorNotHasPaymentMethods)
	}

	var projectPaymentMethodsIds []bson.ObjectId

	for _, pms := range p.PaymentMethods {
		var conPaymentMethod *model.ProjectPaymentModes

		for _, pm := range pms {
			if conPaymentMethod == nil || conPaymentMethod.AddedAt.Before(pm.AddedAt) == true {
				conPaymentMethod = pm
			}
		}

		projectPaymentMethodsIds = append(projectPaymentMethodsIds, conPaymentMethod.Id)
	}

	if len(projectPaymentMethodsIds) <= 0 {
		return nil, errors.New(projectErrorNotHasPaymentMethods)
	}

	projectPaymentMethods := pm.paymentMethodsManager.FindByIds(projectPaymentMethodsIds)

	if projectPaymentMethods == nil || len(projectPaymentMethods) <= 0 {
		return nil, errors.New(projectErrorNotHasPaymentMethods)
	}

	return projectPaymentMethods, nil
}

func (pm *ProjectManager) FindFixedPackage(filters *model.FixedPackageFilters) []*model.FilteredFixedPackage {
	var fps []*model.FilteredFixedPackage

	filters.Region = strings.ToUpper(filters.Region)
	smFps, err := pm.Database.Repository(TableProject).FindFixedPackageByFilters(filters)

	if err != nil {
		pm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableProject, err)
	}

	sFps := smFps[0]["fixed_package"].([]interface{})

	if len(sFps) <= 0 {
		return fps
	}

	for _, v := range sFps {
		vm := v.(map[string]interface{})

		c := pm.currencyManager.FindByCodeInt(vm[model.DBFieldCurrencyInt].(int))

		if c == nil {
			continue
		}

		ffp := &model.FilteredFixedPackage{
			Id:    vm[model.DBFieldId].(string),
			Name:  vm[model.DBFieldName].(string),
			Price: vm[model.DBFieldPrice].(float64),
			Currency: &model.SimpleCurrency{
				Name:    c.Name,
				CodeInt: c.CodeInt,
				CodeA3:  c.CodeA3,
			},
		}

		fps = append(fps, ffp)
	}

	return fps
}

func (pm *ProjectManager) GetProjectsPaymentMethodsByMerchantMainData(mId string) map[string]string {
	projects := pm.FindProjectsByMerchantId(mId, model.DefaultLimit, model.DefaultOffset)

	pms := make(map[string]string)

	if len(projects) <= 0 {
		return pms
	}

	for _, project := range projects {
		projectPms, err := pm.GetProjectPaymentMethods(project.Id)

		if err != nil || len(projectPms) <= 0 {
			continue
		}

		for _, ppm := range projectPms {
			sPpmId := ppm.Id.Hex()

			if _, ok := pms[sPpmId]; ok {
				continue
			}

			pms[sPpmId] = ppm.Name
		}
	}

	return pms
}
