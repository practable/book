package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/icza/gog"
	"github.com/ory/viper"
	apiclient "github.com/practable/book/internal/client/client"
	"github.com/practable/book/internal/client/client/admin"
	"github.com/practable/book/internal/client/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// bookingEditFile is deliberately a small, stable YAML shape. It avoids the
// generated model's YAML field-name quirks and contains only editable booking
// fields; server-owned status must remain absent.
type bookingEditFile struct {
	OriginalName string             `json:"original_name" yaml:"original_name"`
	Revision     int64              `json:"revision" yaml:"revision"`
	Booking      bookingEditBooking `json:"booking" yaml:"booking"`
}

type bookingEditBooking struct {
	Name   string              `json:"name" yaml:"name"`
	Policy string              `json:"policy" yaml:"policy"`
	Slot   string              `json:"slot" yaml:"slot"`
	User   string              `json:"user" yaml:"user"`
	When   bookingEditInterval `json:"when" yaml:"when"`
}

type bookingEditInterval struct {
	Start string `json:"start" yaml:"start"`
	End   string `json:"end" yaml:"end"`
}

func editFileFromModel(edit *models.BookingEdit) (bookingEditFile, error) {
	if edit == nil || edit.OriginalName == nil || edit.Revision == nil || edit.Booking == nil ||
		edit.Booking.Name == nil || edit.Booking.Policy == nil || edit.Booking.Slot == nil || edit.Booking.User == nil || edit.Booking.When == nil {
		return bookingEditFile{}, fmt.Errorf("server returned an incomplete booking edit")
	}
	return bookingEditFile{OriginalName: *edit.OriginalName, Revision: *edit.Revision, Booking: bookingEditBooking{
		Name: *edit.Booking.Name, Policy: *edit.Booking.Policy, Slot: *edit.Booking.Slot, User: *edit.Booking.User,
		When: bookingEditInterval{Start: edit.Booking.When.Start.String(), End: edit.Booking.When.End.String()},
	}}, nil
}

func (f bookingEditFile) toModel() (*models.BookingEdit, error) {
	start, err := time.Parse(time.RFC3339Nano, f.Booking.When.Start)
	if err != nil {
		return nil, fmt.Errorf("parse booking.when.start: %w", err)
	}
	end, err := time.Parse(time.RFC3339Nano, f.Booking.When.End)
	if err != nil {
		return nil, fmt.Errorf("parse booking.when.end: %w", err)
	}
	return &models.BookingEdit{
		OriginalName: gog.Ptr(f.OriginalName), Revision: gog.Ptr(f.Revision),
		Booking: &models.Booking{Name: gog.Ptr(f.Booking.Name), Policy: gog.Ptr(f.Booking.Policy), Slot: gog.Ptr(f.Booking.Slot), User: gog.Ptr(f.Booking.User),
			When: &models.Interval{Start: strfmt.DateTime(start), End: strfmt.DateTime(end)}},
	}, nil
}

func editClient() (*apiclient.Client, runtime.ClientAuthInfoWriter, time.Duration) {
	viper.SetEnvPrefix("BOOK_CLIENT")
	viper.AutomaticEnv()
	viper.SetDefault("host", "localhost")
	viper.SetDefault("scheme", "http")
	viper.SetDefault("base_path", "/api/v1")
	viper.SetDefault("format", "yaml")
	config := apiclient.DefaultTransportConfig().WithHost(viper.GetString("host")).WithSchemes([]string{viper.GetString("scheme")}).WithBasePath(viper.GetString("base_path"))
	return apiclient.NewDiagnosticHTTPClientWithConfig(nil, config), httptransport.APIKeyAuth("Authorization", "header", viper.GetString("token")), 10 * time.Second
}

func requireEditToken() bool {
	if viper.GetString("token") != "" {
		return true
	}
	fmt.Println("BOOK_CLIENT_TOKEN not set")
	return false
}

var bookingsExportOneCmd = &cobra.Command{
	Use:   "export-one BOOKING_NAME",
	Short: "Export one booking in an editable envelope",
	Long: `Exports one unstarted booking together with its revision. Edit the booking
fields and preserve original_name and revision, then use bookings replace-one.

Example: book bookings export-one booking-id > booking-edit.yaml`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !requireEditToken() {
			return
		}
		client, auth, timeout := editClient()
		result, err := client.Admin.ExportBookingForEdit(admin.NewExportBookingForEditParams().WithTimeout(timeout).WithBookingName(args[0]), auth)
		if err != nil {
			fmt.Printf("Error: failed to export booking for editing because %s\n", err)
			return
		}
		file, err := editFileFromModel(result.Payload)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}
		format := strings.ToLower(viper.GetString("format"))
		if format == "json" {
			data, err := json.MarshalIndent(file, "", "  ")
			if err == nil {
				fmt.Println(string(data))
			}
			return
		}
		data, err := yaml.Marshal(file)
		if err != nil {
			fmt.Printf("Error: failed to format booking edit: %s\n", err)
			return
		}
		fmt.Print(string(data))
	},
}

var bookingsReplaceOneCmd = &cobra.Command{
	Use:   "replace-one BOOKING_EDIT_FILE",
	Short: "Atomically replace one unstarted booking",
	Long: `Uploads an envelope created by bookings export-one. The server checks its
revision, validates the edited booking, and atomically supersedes the original.
An exact retry is safe.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !requireEditToken() {
			return
		}
		data, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Printf("Error: failed to read booking edit file: %s\n", err)
			return
		}
		var file bookingEditFile
		format := strings.ToLower(viper.GetString("format"))
		if format == "json" {
			err = json.Unmarshal(data, &file)
		} else {
			err = yaml.Unmarshal(data, &file)
		}
		if err != nil {
			fmt.Printf("Error: failed to parse booking edit file: %s\n", err)
			return
		}
		edit, err := file.toModel()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}
		client, auth, timeout := editClient()
		_, err = client.Admin.ReplaceBooking(admin.NewReplaceBookingParams().WithTimeout(timeout).WithBookingName(file.OriginalName).WithBookingEdit(edit), auth)
		if err != nil {
			fmt.Printf("Error: failed to replace booking because %s\n", err)
		}
	},
}

func init() {
	bookingsCmd.AddCommand(bookingsExportOneCmd, bookingsReplaceOneCmd)
}
