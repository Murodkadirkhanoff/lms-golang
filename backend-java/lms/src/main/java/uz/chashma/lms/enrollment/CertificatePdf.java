package uz.chashma.lms.enrollment;

import com.lowagie.text.Document;
import com.lowagie.text.Element;
import com.lowagie.text.Font;
import com.lowagie.text.PageSize;
import com.lowagie.text.Paragraph;
import com.lowagie.text.Rectangle;
import com.lowagie.text.pdf.PdfWriter;

import java.awt.Color;
import java.io.ByteArrayOutputStream;
import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;

/** Sertifikat PDF'i (OpenPDF): A4 landscape, ramka, ism + kurs + sana. */
final class CertificatePdf {

    private static final DateTimeFormatter DATE = DateTimeFormatter.ofPattern("MMMM d, yyyy")
            .withZone(ZoneOffset.UTC);

    private static final Color INDIGO = new Color(79, 70, 229);
    private static final Color GRAY = new Color(100, 116, 139);

    private CertificatePdf() {
    }

    static byte[] render(String studentName, CertificateRepository.CertificateDto cert) {
        ByteArrayOutputStream out = new ByteArrayOutputStream();

        Document doc = new Document(PageSize.A4.rotate(), 60, 60, 60, 60);
        PdfWriter writer = PdfWriter.getInstance(doc, out);
        doc.open();

        // Ikki qavatli ramka
        Rectangle page = doc.getPageSize();
        Rectangle outer = new Rectangle(30, 30, page.getWidth() - 30, page.getHeight() - 30);
        outer.setBorder(Rectangle.BOX);
        outer.setBorderWidth(3);
        outer.setBorderColor(INDIGO);
        writer.getDirectContent().rectangle(outer);

        Rectangle inner = new Rectangle(40, 40, page.getWidth() - 40, page.getHeight() - 40);
        inner.setBorder(Rectangle.BOX);
        inner.setBorderWidth(1);
        inner.setBorderColor(INDIGO);
        writer.getDirectContent().rectangle(inner);

        Font brand = new Font(Font.HELVETICA, 16, Font.BOLD, INDIGO);
        Font heading = new Font(Font.HELVETICA, 34, Font.BOLD, Color.BLACK);
        Font label = new Font(Font.HELVETICA, 13, Font.NORMAL, GRAY);
        Font name = new Font(Font.HELVETICA, 28, Font.BOLD, INDIGO);
        Font course = new Font(Font.HELVETICA, 22, Font.BOLD, Color.BLACK);
        Font small = new Font(Font.HELVETICA, 11, Font.NORMAL, GRAY);

        doc.add(centered("LearnHub", brand, 10));
        doc.add(centered("Certificate of Completion", heading, 30));
        doc.add(centered("This certificate is proudly presented to", label, 18));
        doc.add(centered(studentName, name, 24));
        doc.add(centered("for successfully completing the course", label, 18));
        doc.add(centered(cert.courseTitle, course, 40));
        doc.add(centered("Issued on %s    •    Certificate ID: LH-%06d"
                .formatted(DATE.format(cert.issuedAt), cert.id), small, 0));

        doc.close();
        return out.toByteArray();
    }

    private static Paragraph centered(String text, Font font, float spacingAfter) {
        Paragraph p = new Paragraph(text, font);
        p.setAlignment(Element.ALIGN_CENTER);
        p.setSpacingAfter(spacingAfter);
        return p;
    }
}
