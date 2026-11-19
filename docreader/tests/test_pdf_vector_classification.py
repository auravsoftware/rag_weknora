import unittest

from docreader.parser.pdf_parser import _classify_page, _is_vector_pdf_document


class TestPDFVectorClassification(unittest.TestCase):
    def test_sandwich_pdf_page_with_invisible_scanner_ocr_is_scanned(self):
        # Scanner embedded an invisible (render-mode 3) English OCR layer
        # ("cho IlQnh viQn tla khoa Phu'ong Ddng") over a cropped scanned image.
        cls = _classify_page(
            image_area_ratio=0.35,
            text_len=450,
            visible_text_objs=0,
            invisible_text_objs=25,
        )
        self.assertEqual(cls, "scanned")

    def test_genuine_vector_page_with_small_logo_is_text(self):
        cls = _classify_page(
            image_area_ratio=0.08,
            text_len=1200,
            visible_text_objs=40,
            invisible_text_objs=0,
        )
        self.assertEqual(cls, "text")

    def test_document_with_95_scanned_and_2_misclassified_text_pages_is_not_vector(self):
        classes = ["scanned"] * 95 + ["text", "text"]
        text_lens = [0] * 95 + [300, 280]
        self.assertFalse(_is_vector_pdf_document(classes, text_lens))

    def test_pure_vector_document_is_vector(self):
        classes = ["text"] * 10
        text_lens = [800] * 10
        self.assertTrue(_is_vector_pdf_document(classes, text_lens))

    def test_mostly_vector_document_with_single_scanned_appendix_remains_vector(self):
        # 9 native vector pages + 1 scanned signature page = 90% vector >= 80%
        classes = ["text"] * 9 + ["scanned"]
        text_lens = [900] * 9 + [0]
        self.assertTrue(_is_vector_pdf_document(classes, text_lens))

    def test_no_vector_content_pages_defaults_to_scanned(self):
        classes = ["scanned", "text"]
        text_lens = [0, 3]
        self.assertFalse(_is_vector_pdf_document(classes, text_lens))


if __name__ == "__main__":
    unittest.main()
