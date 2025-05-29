package ent_impl

import (
	"context"
	"entgo.io/ent/dialect/sql"
	"fmt"
	"github.com/google/uuid"
	"micro-mart/pkg/db"
	"micro-mart/pkg/mmotel"
	"micro-mart/services/transaction-orchestrator/domain/aggregate"
	"micro-mart/services/transaction-orchestrator/domain/entity"
	"micro-mart/services/transaction-orchestrator/domain/model"
	"micro-mart/services/transaction-orchestrator/domain/repository"
	"micro-mart/services/transaction-orchestrator/infrastructure/ent_impl/ent"
	"micro-mart/services/transaction-orchestrator/infrastructure/ent_impl/ent/schema"
	"micro-mart/services/transaction-orchestrator/infrastructure/ent_impl/ent/transaction"
)

type TransactionRepositoryImpl struct {
	db db.Database
}

var _ repository.TransactionRepository = (*TransactionRepositoryImpl)(nil)

func NewTransactionRepositoryImpl(db db.Database) repository.TransactionRepository {
	return &TransactionRepositoryImpl{
		db: db,
	}
}

func (t TransactionRepositoryImpl) SaveTransaction(ctx context.Context, tx *aggregate.Transaction) error {
	// 將聚合轉換為模型以進行儲存
	modelTx := model.FromAggregate(tx)

	// 獲取 Ent 客戶端
	client := t.db.GetClient(ctx).(*ent.Client)

	// 將模型的狀態轉換為 Ent 的狀態類型
	status := transaction.Status(modelTx.Status)

	// 將模型的步驟轉換為 Ent 的步驟類型
	steps := make([]schema.TransactionStep, len(modelTx.Steps))
	for i, step := range modelTx.Steps {
		steps[i] = schema.TransactionStep{
			ID:                  step.ID,
			TransactionID:       step.TransactionID,
			ServiceName:         step.ServiceName,
			Operation:           step.Operation,
			Status:              step.Status,
			Payload:             step.Payload,
			Result:              step.Result,
			CompensationPayload: step.CompensationPayload,
			CreatedAt:           step.CreatedAt,
			UpdatedAt:           step.UpdatedAt,
		}
	}

	// 創建新的交易記錄
	_, err := client.Transaction.Create().
		SetID(tx.ID).
		SetType(modelTx.Type).
		SetStatus(status).
		SetSteps(steps).
		SetCreatedAt(modelTx.CreatedAt).
		SetUpdatedAt(modelTx.UpdatedAt).
		Save(ctx)

	if err != nil {
		// 處理可能的錯誤情況
		if ent.IsConstraintError(err) {
			errMsg := fmt.Sprintf("交易已存在或違反唯一性約束: %v", err)
			mmotel.Error(ctx, errMsg)
			return fmt.Errorf(errMsg)
		}
		errMsg := fmt.Sprintf("儲存交易時發生錯誤: %v", err)
		mmotel.Error(ctx, errMsg)
		return fmt.Errorf(errMsg)
	}

	return nil
}

func (t TransactionRepositoryImpl) UpdateTransaction(ctx context.Context, tx *aggregate.Transaction) error {
	// 將聚合轉換為模型以進行儲存
	modelTx := model.FromAggregate(tx)

	// 獲取 Ent 客戶端
	client := t.db.GetClient(ctx).(*ent.Client)

	// 將模型的狀態轉換為 Ent 的狀態類型
	status := transaction.Status(modelTx.Status)

	// 將模型的步驟轉換為 Ent 的步驟類型
	steps := make([]schema.TransactionStep, len(modelTx.Steps))
	for i, step := range modelTx.Steps {
		steps[i] = schema.TransactionStep{
			ID:                  step.ID,
			TransactionID:       step.TransactionID,
			ServiceName:         step.ServiceName,
			Operation:           step.Operation,
			Status:              step.Status,
			Payload:             step.Payload,
			Result:              step.Result,
			CompensationPayload: step.CompensationPayload,
			CreatedAt:           step.CreatedAt,
			UpdatedAt:           step.UpdatedAt,
		}
	}

	// 更新現有的交易記錄
	_, err := client.Transaction.UpdateOneID(tx.ID).
		SetType(modelTx.Type).
		SetStatus(status).
		SetSteps(steps).
		SetUpdatedAt(modelTx.UpdatedAt).
		Save(ctx)

	if err != nil {
		// 處理可能的錯誤情況
		if ent.IsNotFound(err) {
			errMsg := fmt.Sprintf("找不到 ID 為 %s 的交易", tx.ID.String())
			mmotel.Error(ctx, errMsg)
			return fmt.Errorf(errMsg)
		}
		if ent.IsConstraintError(err) {
			errMsg := fmt.Sprintf("更新交易時違反唯一性約束: %v", err)
			mmotel.Error(ctx, errMsg)
			return fmt.Errorf(errMsg)
		}
		errMsg := fmt.Sprintf("更新交易時發生錯誤: %v", err)
		mmotel.Error(ctx, errMsg)
		return fmt.Errorf(errMsg)
	}

	return nil
}

func (t TransactionRepositoryImpl) GetTransaction(ctx context.Context, id string) (*aggregate.Transaction, error) {
	// 獲取 Ent 客戶端
	client := t.db.GetClient(ctx).(*ent.Client)

	// 解析 ID 字串為 UUID
	uuid, err := uuid.Parse(id)
	if err != nil {
		// 記錄錯誤：無效的 UUID 格式
		errMsg := fmt.Sprintf("無效的交易 ID 格式: %v", err)
		mmotel.Error(ctx, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// 從資料庫中查詢交易
	entTx, err := client.Transaction.Get(ctx, uuid)
	if err != nil {
		// 處理未找到的情況
		if ent.IsNotFound(err) {
			errMsg := fmt.Sprintf("找不到 ID 為 %s 的交易", id)
			mmotel.Error(ctx, errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		// 記錄其他資料庫錯誤
		errMsg := fmt.Sprintf("查詢交易時發生錯誤: %v", err)
		mmotel.Error(ctx, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// 將 Ent 交易轉換為模型
	modelTx := &model.Transaction{
		ID:        entTx.ID.String(),
		Type:      entTx.Type,
		Status:    model.TransactionStatus(entTx.Status),
		Steps:     make([]model.TransactionStep, len(entTx.Steps)),
		CreatedAt: entTx.CreatedAt,
		UpdatedAt: entTx.UpdatedAt,
	}

	// 轉換步驟
	for i, step := range entTx.Steps {
		modelTx.Steps[i] = model.TransactionStep{
			ID:                  step.ID,
			TransactionID:       step.TransactionID,
			ServiceName:         step.ServiceName,
			Operation:           step.Operation,
			Status:              step.Status,
			Payload:             step.Payload,
			Result:              step.Result,
			CompensationPayload: step.CompensationPayload,
			CreatedAt:           step.CreatedAt,
			UpdatedAt:           step.UpdatedAt,
		}
	}

	// 將模型轉換為聚合並返回
	return modelTx.ToAggregate(), nil
}

func (t TransactionRepositoryImpl) FindTransactionsByStatus(ctx context.Context, statuses []entity.TransactionStatus) ([]*aggregate.Transaction, error) {
	// 獲取 Ent 客戶端
	client := t.db.GetClient(ctx).(*ent.Client)

	// 將實體狀態轉換為 Ent 狀態
	entStatuses := make([]transaction.Status, len(statuses))
	for i, status := range statuses {
		entStatuses[i] = transaction.Status(status)
	}

	// 查詢符合狀態的交易
	query := client.Transaction.Query()

	// 如果有狀態過濾條件，則添加到查詢中
	if len(entStatuses) > 0 {
		predicates := make([]func(*sql.Selector), len(entStatuses))
		for i, status := range entStatuses {
			// 使用閉包捕獲每個狀態值
			s := status // 創建本地變數以避免閉包問題
			predicates[i] = func(selector *sql.Selector) {
				selector.Where(sql.EQ(transaction.FieldStatus, s))
			}
		}

		// 使用 Or 組合所有狀態條件
		query = query.Where(func(selector *sql.Selector) {
			for _, predicate := range predicates {
				predicate(selector)
			}
		})
	}

	// 執行查詢
	entTxs, err := query.All(ctx)
	if err != nil {
		errMsg := fmt.Sprintf("查詢交易時發生錯誤: %v", err)
		mmotel.Error(ctx, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// 如果沒有找到符合條件的交易，返回空切片
	if len(entTxs) == 0 {
		return []*aggregate.Transaction{}, nil
	}

	// 將 Ent 交易轉換為聚合
	aggregateTxs := make([]*aggregate.Transaction, len(entTxs))
	for i, entTx := range entTxs {
		// 將 Ent 交易轉換為模型
		modelTx := &model.Transaction{
			ID:        entTx.ID.String(),
			Type:      entTx.Type,
			Status:    model.TransactionStatus(entTx.Status),
			Steps:     make([]model.TransactionStep, len(entTx.Steps)),
			CreatedAt: entTx.CreatedAt,
			UpdatedAt: entTx.UpdatedAt,
		}

		// 轉換步驟
		for j, step := range entTx.Steps {
			modelTx.Steps[j] = model.TransactionStep{
				ID:                  step.ID,
				TransactionID:       step.TransactionID,
				ServiceName:         step.ServiceName,
				Operation:           step.Operation,
				Status:              step.Status,
				Payload:             step.Payload,
				Result:              step.Result,
				CompensationPayload: step.CompensationPayload,
				CreatedAt:           step.CreatedAt,
				UpdatedAt:           step.UpdatedAt,
			}
		}

		// 將模型轉換為聚合
		aggregateTxs[i] = modelTx.ToAggregate()
	}

	return aggregateTxs, nil
}

func (t TransactionRepositoryImpl) DeleteTransaction(ctx context.Context, id string) error {
	// 獲取 Ent 客戶端
	client := t.db.GetClient(ctx).(*ent.Client)

	// 解析 ID 字串為 UUID
	uuid, err := uuid.Parse(id)
	if err != nil {
		errMsg := fmt.Sprintf("無效的交易 ID 格式: %v", err)
		mmotel.Error(ctx, errMsg)
		return fmt.Errorf(errMsg)
	}

	// 刪除交易
	err = client.Transaction.DeleteOneID(uuid).Exec(ctx)
	if err != nil {
		// 處理可能的錯誤情況
		if ent.IsNotFound(err) {
			errMsg := fmt.Sprintf("找不到 ID 為 %s 的交易", id)
			mmotel.Error(ctx, errMsg)
			return fmt.Errorf(errMsg)
		}
		errMsg := fmt.Sprintf("刪除交易時發生錯誤: %v", err)
		mmotel.Error(ctx, errMsg)
		return fmt.Errorf(errMsg)
	}

	return nil
}
