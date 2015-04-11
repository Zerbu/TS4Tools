import QtQuick 2.4
import QtQuick.Controls 1.3

Item {
	Column {
		spacing: 4

		Label { text: "Type string to hash:" }

		TextField {
			width: parent.width
			onTextChanged: {
				hash.changeText(text)
			}
		}

		Row {
			spacing: 4

			Label {
				id: formatLabel
				text: "Display result as"
			}

			ExclusiveGroup { id: numberFormat }

			RadioButton {
				text: "Hexadecimal"
				checked: true
				exclusiveGroup: numberFormat
				anchors.baseline: formatLabel.baseline
				onCheckedChanged: {
					if (checked) {
						hash.changeFormat("hex")
					}
				}
			}

			RadioButton {
				text: "Decimal"
				exclusiveGroup: numberFormat
				anchors.baseline: formatLabel.baseline
				onCheckedChanged: {
					if (checked) {
						hash.changeFormat("dec")
					}
				}
			}
		}

		Grid {
			columns: 2
			spacing: 4

			Column {
				spacing: 4

				Label { text: "FNV 24" }

				TextField {
					text: hash.fnv24
					width: 160
					readOnly: true
					horizontalAlignment: TextInput.AlignRight
				}
			}

			Item { width: 1; height: 1 }

			Column {
				spacing: 4

				Label { text: "FNV 32" }

				TextField {
					text: hash.fnv32
					width: 160
					readOnly: true
					horizontalAlignment: TextInput.AlignRight
				}
			}

			Column {
				spacing: 4

				Label { text: "FNV 32 High Bit" }

				TextField {
					text: hash.fnv32
					width: 160
					readOnly: true
					horizontalAlignment: TextInput.AlignRight
				}
			}

			Column {
				spacing: 4

				Label { text: "FNV 64" }

				TextField {
					text: hash.fnv64
					width: 160
					readOnly: true
					horizontalAlignment: TextInput.AlignRight
				}
			}

			Column {
				spacing: 4

				Label { text: "FNV 64 High Bit" }

				TextField {
					text: hash.fnv64High
					width: 160
					readOnly: true
					horizontalAlignment: TextInput.AlignRight
				}
			}
		}
	}
}
