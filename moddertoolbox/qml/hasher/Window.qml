import QtQuick 2.4
import QtQuick.Controls 1.3

ApplicationWindow {
	property real windowMargin: 8
	property real windowSpacing: 4
	property int hashFieldWidth: 160

	title: "Hasher"
	width: body.width + 2 * windowMargin + 2
	height: body.height + 2 * windowMargin + 2
	minimumWidth: width
	minimumHeight: height
	maximumWidth: width
	maximumHeight: height

	Column {
		id: body
		spacing: windowSpacing
		anchors.top: parent.top
		anchors.left: parent.left
		anchors.margins: windowMargin

		Label { text: "Type string to hash:" }

		TextField {
			width: parent.width
			onTextChanged: { app.changeText(text) }
		}

		Row {
			spacing: windowSpacing

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
						app.changeFormat("hex")
					}
				}
			}

			RadioButton {
				text: "Decimal"
				exclusiveGroup: numberFormat
				anchors.baseline: formatLabel.baseline
				onCheckedChanged: {
					if (checked) {
						app.changeFormat("dec")
					}
				}
			}
		}

		Grid {
			columns: 2
			spacing: windowSpacing

			Column {
				spacing: windowSpacing

				Label { text: "FNV 24" }

				TextField {
					text: app.fnv24
					width: hashFieldWidth
					readOnly: true
					horizontalAlignment: TextInput.AlignRight
				}
			}

			Item { width: 1; height: 1 }

			Column {
				spacing: windowSpacing

				Label { text: "FNV 32" }

				TextField {
					text: app.fnv32
					width: hashFieldWidth
					readOnly: true
					horizontalAlignment: TextInput.AlignRight
				}
			}

			Column {
				spacing: windowSpacing

				Label { text: "FNV 32 High Bit" }

				TextField {
					text: app.fnv32High
					width: hashFieldWidth
					readOnly: true
					horizontalAlignment: TextInput.AlignRight
				}
			}

			Column {
				spacing: windowSpacing

				Label { text: "FNV 64" }

				TextField {
					text: app.fnv64
					width: hashFieldWidth
					readOnly: true
					horizontalAlignment: TextInput.AlignRight
				}
			}

			Column {
				spacing: windowSpacing

				Label { text: "FNV 64 High Bit" }

				TextField {
					text: app.fnv64High
					width: hashFieldWidth
					readOnly: true
					horizontalAlignment: TextInput.AlignRight
				}
			}
		}
	}
}
