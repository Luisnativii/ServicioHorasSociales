package reportes

import (
	"archive/zip"
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
)

const xmlHeader = "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
const mainNS = "http://schemas.openxmlformats.org/spreadsheetml/2006/main"
const relNS = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
const pkgRelNS = "http://schemas.openxmlformats.org/package/2006/relationships"

func WriteXLSX(out io.Writer, report Report, filter Filter) error {
	z := zip.NewWriter(out)
	for _, part := range []struct{ path, data string }{
		{"[Content_Types].xml", `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/><Override PartName="/xl/worksheets/sheet2.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`},
		{"_rels/.rels", `<Relationships xmlns="` + pkgRelNS + `"><Relationship Id="rId1" Type="` + relNS + `/officeDocument" Target="xl/workbook.xml"/></Relationships>`},
		{"xl/workbook.xml", `<workbook xmlns="` + mainNS + `" xmlns:r="` + relNS + `"><sheets><sheet name="Resumen" sheetId="1" r:id="rId1"/><sheet name="Movimientos" sheetId="2" r:id="rId2"/></sheets></workbook>`},
		{"xl/_rels/workbook.xml.rels", `<Relationships xmlns="` + pkgRelNS + `"><Relationship Id="rId1" Type="` + relNS + `/worksheet" Target="worksheets/sheet1.xml"/><Relationship Id="rId2" Type="` + relNS + `/worksheet" Target="worksheets/sheet2.xml"/></Relationships>`},
	} {
		w, err := z.Create(part.path)
		if err != nil {
			return err
		}
		if _, err = io.WriteString(w, xmlHeader+part.data); err != nil {
			return err
		}
	}

	w, err := startSheet(z, "xl/worksheets/sheet1.xml")
	if err != nil {
		return err
	}
	for i, row := range [][]any{
		{"Indicador", "Valor"},
		{"Estudiantes únicos", report.Summary.Students},
		{"Participaciones", report.Summary.Participations},
		{"Proyectos", report.Summary.Projects},
		{"Horas internas", report.Summary.Internal},
		{"Horas externas", report.Summary.External},
		{"Cumplimientos con fecha conocida", report.Summary.Completed},
		{"Movimientos", report.Total},
		{"Fecha de actividad desde", filter.From},
		{"Fecha de actividad hasta", filter.To},
		{"Estado vigente", filter.State},
		{"Departamento vigente (UUID)", filter.Department},
		{"Municipio vigente (UUID)", filter.Municipality},
		{"Distrito vigente (UUID)", filter.District},
		{"Carrera histórica (UUID)", filter.Career},
		{"Tipo histórico (UUID)", filter.Type},
	} {
		if err = writeRow(w, i+1, row...); err != nil {
			return err
		}
	}
	for i, area := range filter.Areas {
		if err = writeRow(w, 17+i, "Área histórica (UUID)", area); err != nil {
			return err
		}
	}
	if err = finishSheet(w); err != nil {
		return err
	}
	w, err = startSheet(z, "xl/worksheets/sheet2.xml")
	if err != nil {
		return err
	}
	if err = writeRow(w, 1, "Carnet", "Estudiante", "Proyecto", "Fecha de actividad", "Horas", "Naturaleza", "Facultad histórica",
		"Carrera histórica", "Tipo histórico", "Área histórica", "Área CSS original", "Distrito vigente", "Estado vigente"); err != nil {
		return err
	}
	for i, row := range report.Rows {
		if err = writeRow(w, i+2, row.Carnet, row.Student, row.Project, row.Date, int64(row.Hours), row.Nature, history(row.Faculty),
			history(row.Career), history(row.Type), history(row.Area), optional(row.LegacyArea), optional(row.District), row.State); err != nil {
			return err
		}
	}
	if err = finishSheet(w); err != nil {
		return err
	}
	return z.Close()
}

func history(value *string) string {
	if value == nil {
		return "Sin dato histórico"
	}
	return *value
}

func optional(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func startSheet(z *zip.Writer, path string) (*bufio.Writer, error) {
	out, err := z.Create(path)
	if err != nil {
		return nil, err
	}
	w := bufio.NewWriter(out)
	_, err = w.WriteString(xmlHeader + `<worksheet xmlns="` + mainNS + `"><sheetData>`)
	return w, err
}

func finishSheet(w *bufio.Writer) error {
	if _, err := w.WriteString("</sheetData></worksheet>"); err != nil {
		return err
	}
	return w.Flush()
}

func writeRow(w *bufio.Writer, index int, fields ...any) error {
	number := strconv.Itoa(index)
	if _, err := w.WriteString(`<row r="` + number + `">`); err != nil {
		return err
	}
	for i, field := range fields {
		reference := string(rune('A'+i)) + number
		switch v := field.(type) {
		case string:
			// Even leading formula characters are stored verbatim as inline text.
			if _, err := w.WriteString(`<c r="` + reference + `" t="inlineStr"><is><t xml:space="preserve">`); err != nil {
				return err
			}
			if err := xml.EscapeText(w, []byte(v)); err != nil {
				return err
			}
			if _, err := w.WriteString("</t></is></c>"); err != nil {
				return err
			}
		case int64:
			if _, err := w.WriteString(`<c r="` + reference + `" t="n"><v>` + strconv.FormatInt(v, 10) + `</v></c>`); err != nil {
				return err
			}
		default:
			return fmt.Errorf("tipo de celda no admitido: %T", field)
		}
	}
	_, err := w.WriteString("</row>")
	return err
}
