package template

import (
	"github.com/spf13/cobra"
)

// Cmd is the base command for template management: `structify template`.
var Cmd = &cobra.Command{
	Use:   "template",
	Short: "Gestionar templates de Structify",
	Long: "Administra templates locales, built-in y remotos.\n\n" +
		"Incluye comandos para listar, instalar, importar, crear, editar,\n" +
		"validar, eliminar, inspeccionar, actualizar y preparar publicacion.\n\n" +
		"Ejemplos:\n" +
		"  structify template list\n" +
		"  structify template add github.com/user/repo\n" +
		"  structify template import ./mi-proyecto",
}
