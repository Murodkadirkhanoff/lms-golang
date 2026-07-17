package uz.chashma.lms.enrollment;

import org.springframework.jdbc.core.RowCallbackHandler;
import org.springframework.jdbc.core.RowMapper;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDate;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

/** Go enrollment-service/internal/data/orders.go porti. */
@Repository
class OrderRepository {

    private final JdbcClient jdbc;

    OrderRepository(JdbcClient jdbc) {
        this.jdbc = jdbc;
    }

    private static final RowMapper<OrderDto> MAPPER = (rs, rowNum) -> {
        OrderDto o = new OrderDto();
        o.dbId = rs.getLong("id");
        o.createdAt = rs.getObject("created_at", OffsetDateTime.class).toInstant();
        o.userId = rs.getLong("user_id");
        o.total = rs.getDouble("total_amount");
        o.paymentMethod = rs.getString("payment_method");
        o.status = rs.getString("status");
        return o;
    };

    /** Buyurtma va itemlarni bitta tranzaksiyada yozadi (total — DB trigger). */
    @Transactional
    void insert(OrderDto order) {
        jdbc.sql("""
                INSERT INTO enrollment.orders (user_id, status, payment_method)
                VALUES (:userId, :status, :paymentMethod)
                RETURNING id, created_at
                """)
                .param("userId", order.userId)
                .param("status", order.status)
                .param("paymentMethod", order.paymentMethod)
                .query((RowCallbackHandler) rs -> {
                    order.dbId = rs.getLong("id");
                    order.createdAt = rs.getObject("created_at", OffsetDateTime.class).toInstant();
                });

        for (OrderItemDto item : order.items) {
            jdbc.sql("""
                    INSERT INTO enrollment.order_items
                        (order_id, course_id, lesson_id, title, instructor_name, thumbnail_color, price)
                    VALUES (:orderId, :courseId, :lessonId, :title, :instructor, :thumbnailColor, :price)
                    """)
                    .param("orderId", order.dbId)
                    .param("courseId", item.courseId)
                    .param("lessonId", item.lessonId)
                    .param("title", item.courseTitle)
                    .param("instructor", item.instructor)
                    .param("thumbnailColor", item.thumbnailColor)
                    .param("price", item.price)
                    .update();
            order.total += item.price;
        }
    }

    Optional<OrderDto> findForUser(long id, long userId) {
        if (id < 1) {
            return Optional.empty();
        }

        Optional<OrderDto> order = jdbc.sql("""
                SELECT id, created_at, user_id, total_amount, payment_method, status
                FROM enrollment.orders
                WHERE id = :id AND user_id = :userId
                """)
                .param("id", id)
                .param("userId", userId)
                .query(MAPPER)
                .optional();

        order.ifPresent(o -> {
            o.items = itemsForOrders(List.of(o.dbId)).getOrDefault(o.dbId, new ArrayList<>());
            o.finalizeView();
        });

        return order;
    }

    record OrderPage(List<OrderDto> orders, int total) {
    }

    OrderPage listByUser(long userId, int page, int pageSize) {
        var totalHolder = new int[1];
        List<OrderDto> orders = jdbc.sql("""
                SELECT count(*) OVER() AS total, id, created_at, user_id, total_amount, payment_method, status
                FROM enrollment.orders
                WHERE user_id = :userId
                ORDER BY created_at DESC, id DESC
                LIMIT :limit OFFSET :offset
                """)
                .param("userId", userId)
                .param("limit", pageSize)
                .param("offset", (page - 1) * pageSize)
                .query((rs, rowNum) -> {
                    totalHolder[0] = rs.getInt("total");
                    return MAPPER.mapRow(rs, rowNum);
                })
                .list();

        if (orders.isEmpty()) {
            return new OrderPage(orders, totalHolder[0]);
        }

        Map<Long, List<OrderItemDto>> items = itemsForOrders(
                orders.stream().map(o -> o.dbId).toList());

        for (OrderDto order : orders) {
            order.items = items.getOrDefault(order.dbId, new ArrayList<>());
            order.finalizeView();
        }

        return new OrderPage(orders, totalHolder[0]);
    }

    /** Barcha to'langan buyurtmalar summasi (admin stats). */
    double revenue() {
        Double revenue = jdbc.sql("SELECT COALESCE(sum(total_amount), 0) FROM enrollment.orders WHERE status = 'paid'")
                .query(Double.class)
                .single();
        return revenue == null ? 0 : revenue;
    }

    record MonthRevenue(String month, double revenue) {
    }

    record RevenueResult(double total, List<MonthRevenue> monthly) {
    }

    /**
     * Instruktor kurslari (va darslari) bo'yicha daromad: jami va so'nggi
     * 6 oyning oylik taqsimoti (bo'sh oylar 0 bilan to'ldiriladi).
     */
    RevenueResult revenueForItems(List<Long> courseIds, List<Long> lessonIds) {
        Map<String, Double> byMonth = new HashMap<>();
        double[] total = {0};

        // Bo'sh ro'yxatlar uchun IN () sintaksis xatosidan qochamiz.
        List<Long> safeCourseIds = courseIds.isEmpty() ? List.of(-1L) : courseIds;
        List<Long> safeLessonIds = lessonIds.isEmpty() ? List.of(-1L) : lessonIds;

        jdbc.sql("""
                SELECT to_char(date_trunc('month', o.created_at), 'YYYY-MM') AS month,
                       COALESCE(sum(oi.price), 0) AS revenue
                FROM enrollment.order_items oi
                JOIN enrollment.orders o ON o.id = oi.order_id
                WHERE o.status = 'paid'
                  AND (oi.course_id IN (:courseIds) OR oi.lesson_id IN (:lessonIds))
                GROUP BY month
                """)
                .param("courseIds", safeCourseIds)
                .param("lessonIds", safeLessonIds)
                .query((RowCallbackHandler) rs -> {
                    double revenue = rs.getDouble("revenue");
                    byMonth.put(rs.getString("month"), revenue);
                    total[0] += revenue;
                });

        DateTimeFormatter monthFormat = DateTimeFormatter.ofPattern("yyyy-MM");
        LocalDate firstOfMonth = LocalDate.now(ZoneOffset.UTC).withDayOfMonth(1);

        List<MonthRevenue> monthly = new ArrayList<>(6);
        for (int i = 5; i >= 0; i--) {
            String month = firstOfMonth.minusMonths(i).format(monthFormat);
            monthly.add(new MonthRevenue(month, byMonth.getOrDefault(month, 0.0)));
        }

        return new RevenueResult(total[0], monthly);
    }

    private Map<Long, List<OrderItemDto>> itemsForOrders(List<Long> orderIds) {
        Map<Long, List<OrderItemDto>> byOrder = new LinkedHashMap<>();
        jdbc.sql("""
                SELECT order_id, course_id, lesson_id, title, instructor_name, thumbnail_color, price
                FROM enrollment.order_items
                WHERE order_id IN (:ids)
                ORDER BY id
                """)
                .param("ids", orderIds)
                .query((RowCallbackHandler) rs -> {
                    OrderItemDto item = new OrderItemDto();
                    long courseId = rs.getLong("course_id");
                    item.courseId = rs.wasNull() ? null : courseId;
                    long lessonId = rs.getLong("lesson_id");
                    item.lessonId = rs.wasNull() ? null : lessonId;
                    item.courseTitle = rs.getString("title");
                    item.instructor = rs.getString("instructor_name");
                    item.thumbnailColor = rs.getString("thumbnail_color");
                    item.price = rs.getDouble("price");
                    byOrder.computeIfAbsent(rs.getLong("order_id"), k -> new ArrayList<>()).add(item);
                });
        return byOrder;
    }
}
