# Sample Import Fixtures

This directory contains sample CSV files for testing imports in the Lakehouse Restaurant Finance application.

## Files

### sample-pos.csv

Point of Sale / Sales transaction data with the following columns:

- **Date**: Transaction date (YYYY-MM-DD)
- **Time**: Transaction time (HH:MM)
- **Total**: Total amount including tax
- **Subtotal**: Amount before tax
- **Tax**: GST amount (10% in Australia)
- **Discounts**: Any discounts applied
- **Comps**: Complimentary items value
- **Payment Method**: Card, Cash, etc.
- **Channel**: Dine In, Takeaway, Pickup, Catering
- **Server**: Staff member name

### sample-payroll.csv

Payroll / labor cost data with the following columns:

- **Period Start**: Pay period start date
- **Period End**: Pay period end date
- **Employee**: Employee name
- **Hours Worked**: Total hours in period
- **Hourly Rate**: Rate per hour (AUD)
- **Total Wages**: Gross wages
- **Super**: Superannuation contribution (11% in 2024)
- **Tax Withheld**: PAYG withholding

### sample-inventory.csv

Inventory snapshot data with the following columns:

- **Snapshot Date**: Date of inventory count
- **Item Name**: Product name
- **Category**: Beverages, Dairy, Meat, Seafood, Produce, Pantry, Bakery
- **Quantity**: Amount on hand
- **Unit**: Unit of measure (kg, L, ea, loaf, bottle)
- **Unit Cost**: Cost per unit (AUD)
- **Total Value**: Quantity × Unit Cost

## Usage

1. Navigate to the Imports page
2. Select the appropriate data source type
3. Choose a mapping profile or use defaults
4. Upload the CSV file
5. Review import results and any anomalies

## Expected Results

After importing all three sample files:

- **Revenue**: ~$2,100 from 24 transactions
- **Labor Cost**: ~$10,700 over 2 pay periods
- **Inventory Value**: ~$2,400

## Column Mapping

The default mappings match these column headers. If your POS/payroll/inventory exports use different column names, create a custom mapping profile in the application.
