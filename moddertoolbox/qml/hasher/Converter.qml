import QtQuick 2.4
import QtQuick.Controls 1.3

Item {
	Column {
		spacing: 4

		Label { text: "Enter number to convert to hexadecimal or decimal: " }

		TextField {
			width: parent.width
			onTextChanged: {
				convert.changeText(text)
			}
		}

		Label { text: "Result:" }

		TextField {
			text: convert.result
			width: parent.width
			readOnly: true
			onTextChanged: {
				if (convert.alignRight) {
					horizontalAlignment = TextInput.AlignRight
				} else {
					horizontalAlignment = TextInput.AlignLeft
				}
			}
		}
	}
}
