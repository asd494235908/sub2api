UPDATE users
SET total_recharged = COALESCE(recharged.total_pay_amount, 0)
FROM (
  SELECT
    u.id AS user_id,
    COALESCE(SUM(
      CASE
        WHEN po.status = 'COMPLETED' THEN po.pay_amount
        WHEN po.status = 'PARTIALLY_REFUNDED' THEN GREATEST(po.pay_amount - COALESCE(po.refund_amount, 0) * po.pay_amount / NULLIF(po.amount, 0), 0)
        ELSE 0
      END
    ), 0) AS total_pay_amount
  FROM users u
  LEFT JOIN payment_orders po
    ON po.user_id = u.id
   AND po.order_type IN ('balance', 'subscription')
   AND po.status IN ('COMPLETED', 'PARTIALLY_REFUNDED')
  GROUP BY u.id
) AS recharged
WHERE users.id = recharged.user_id;
