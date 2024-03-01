set_shepherd_binary_path() {
	local os=$(uname)
	local arch=$(uname -m)
	local binary_name=""

	case $os in
	Darwin)
		case $arch in
		x86_64) binary_name="shepherd-darwin-amd64" ;;
		arm64) binary_name="shepherd-darwin-arm64" ;;
		*)
			echo "Unsupported architecture: $arch"
			exit 1
			;;
		esac
		;;
	Linux)
		case $arch in
		x86_64) binary_name="shepherd-linux-amd64" ;;
		aarch64) binary_name="shepherd-linux-arm64" ;;
		*)
			echo "Unsupported architecture: $arch"
			exit 1
			;;
		esac
		;;
	*)
		echo "Unsupported operating system: $os"
		exit 1
		;;
	esac

	if [ -f "$PWD/src/binaries/$binary_name" ]; then
		chmod +x "$PWD/src/binaries/$binary_name"
		export SHEPHERD_BINARY_PATH="$PWD/src/binaries/$binary_name"
	else
		export SHEPHERD_BINARY_PATH=$(which shepherd)
	fi
}
