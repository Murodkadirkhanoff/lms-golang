package uz.chashma.lms.enrollment;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.time.Instant;
import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.List;

/**
 * Frontend Order tipi bilan bir xil JSON: id string, date, status
 * (completed|refunded|pending), paymentMethod, items, total.
 */
class OrderDto {

    private static final DateTimeFormatter DATE = DateTimeFormatter.ofPattern("yyyy-MM-dd")
            .withZone(ZoneOffset.UTC);

    @JsonIgnore
    public long dbId;
    @JsonIgnore
    public Instant createdAt;
    @JsonIgnore
    public long userId;
    @JsonIgnore
    public String status;

    public String paymentMethod;
    public List<OrderItemDto> items = new ArrayList<>();
    public double total;

    // Frontend uchun hisoblangan ko'rinish maydonlari (Go Order.Finalize).
    @JsonProperty("id")
    public String publicId;
    public String date;
    @JsonProperty("status")
    public String publicStatus;

    void finalizeView() {
        publicId = Long.toString(dbId);
        date = DATE.format(createdAt);

        publicStatus = switch (status) {
            case "paid" -> "completed";
            case "failed", "cancelled" -> "refunded";
            default -> "pending";
        };
    }
}
