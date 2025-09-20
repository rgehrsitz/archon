# CSV Import Plugin

A comprehensive example plugin that demonstrates how to create an importer for Archon. This plugin converts CSV files into hierarchical node structures suitable for Archon projects.

## Features

- **CSV Parsing**: Handles quoted fields, escaped characters, and various CSV formats
- **Hierarchical Mapping**: Groups data by category columns to create meaningful hierarchies
- **Data Type Inference**: Automatically detects numbers, booleans, dates, and strings
- **Flexible Configuration**: Supports custom column mappings and grouping options
- **Error Handling**: Robust error handling with detailed error messages

## Plugin Structure

```
csv-import/
├── archon-plugin.json    # Plugin manifest
├── index.js             # Main plugin implementation
├── sample-data.csv      # Example CSV file
└── README.md           # This file
```

## Usage

### Basic Import

```javascript
// Import CSV data as flat structure
const result = await csvImporter.run(csvData);
// Creates: CSV Import -> Row 1, Row 2, Row 3, ...
```

### Hierarchical Import

```javascript
// Import with category grouping
const result = await csvImporter.run(csvData, {
  hierarchyColumn: 'category',
  nameColumn: 'name'
});
// Creates: CSV Import -> Electronics -> Widget A, Widget B, ...
//                    -> Tools -> Gadget X, Gadget Y, ...
//                    -> Books -> Book Alpha, Book Beta, ...
```

## Configuration Options

- `hierarchyColumn`: Column name to use for grouping (default: 'category')
- `nameColumn`: Column name to use for node names (default: 'name')

## Data Type Inference

The plugin automatically infers data types from CSV values:

- **Numbers**: `123`, `45.67` → `number`
- **Booleans**: `true`, `false`, `yes`, `no`, `1`, `0` → `boolean`
- **Dates**: Valid date strings → `string` (preserved as-is)
- **Strings**: Everything else → `string`

## Example Output

Given the sample CSV data, the plugin creates:

```
CSV Import
├── Electronics
│   ├── Widget A (price: 29.99, stock: 150, active: true)
│   ├── Widget B (price: 45.50, stock: 75, active: true)
│   ├── Cable USB-C (price: 12.99, stock: 300, active: true)
│   └── Power Adapter (price: 18.99, stock: 120, active: true)
├── Tools
│   ├── Gadget X (price: 89.99, stock: 25, active: true)
│   ├── Gadget Y (price: 34.99, stock: 100, active: true)
│   └── Screwdriver Set (price: 67.99, stock: 50, active: true)
└── Books
    ├── Book Alpha (price: 24.99, stock: 200, active: true)
    ├── Book Beta (price: 19.99, stock: 150, active: true)
    └── Manual Guide (price: 9.99, stock: 500, active: true)
```

## Plugin Development Notes

This plugin demonstrates several key concepts:

1. **Plugin Structure**: Proper manifest and implementation structure
2. **Error Handling**: Comprehensive error handling with PluginError
3. **Data Processing**: Complex data transformation and validation
4. **Type Safety**: Type inference and validation
5. **Hierarchical Creation**: Building tree structures from flat data
6. **Configuration**: Flexible options for different use cases

## Testing

Use the included `sample-data.csv` file to test the plugin:

```bash
# Test with sample data
node -e "
import csvImporter from './index.js';
import fs from 'fs';
const csvData = fs.readFileSync('./sample-data.csv', 'utf8');
csvImporter.run(csvData, { hierarchyColumn: 'category' })
  .then(result => console.log(JSON.stringify(result, null, 2)))
  .catch(err => console.error(err));
"
```

## Extending the Plugin

This plugin can be extended with:

- **Additional Data Sources**: Support for Excel, JSON, XML
- **Advanced Mapping**: Custom field mapping and transformation rules
- **Validation**: Data validation and cleaning
- **Batch Processing**: Handle large files efficiently
- **Custom Hierarchies**: More complex grouping strategies
