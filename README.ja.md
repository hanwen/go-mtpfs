## イントロダクション

![Diagram](./img/totemo_wakariyasui_zu.jpg)

**English version is [here](./README.md)**

mtplvcapは、NikonのカメラのLive ViewをWebSocketでブラウザにリレーするマルチプラットフォーム対応 (Windows, Mac, Linux) のソフトです。

mtplvcapとOBSを組み合わせることで、NikonのカメラをHDMIキャプチャーボードなしでWebカメラにできます。お気に入りのカメラでGoogle Hangouts・Meet・Zoomを楽しみましょう！


## 動作を確認しているもの

カメラ:
 - Nikon D5300
 - Nikon D3300 (thanks @unasuke)
 - フィードバック歓迎！お手持ちのカメラの動作可否を是非Issueにあげてください。

OS:
 - Windows 10 version 1909, OS build 18363.900, MSYS2 (MinGW x86_64), amd64
 - macOS 10.15.5 Catalina, amd64
 - Debian GNU/Linux 10 Buster, amd64


## インストール方法

注釈: このセクションにあるShellのコードスニペットは全部コピペできます。


### Windows

**重要！ Windowsでは、カメラを接続して自動でインストールされるドライバをlibusbの汎用ドライバで置き換える必要があります。
置き換えると、手動でドライバを指定し直さない限りカメラはMTPデバイスとして認識されなくなります。以上を認識の上で続行してください。**


#### 1. 一眼レフのドライバを置き換える

1. カメラをPCに接続します
1. Zadigを[ここ](https://zadig.akeo.ie/)からダウンロードして開きます
1. `List All Devices` にチェックを入れます

    <img alt="List All Devices にチェックを入れる" src="./img/zadig_1.png" width="400px">

1. 画面上のリストにカメラ名があるのを確認し、選択します

    <img alt="リストからカメラを選ぶ" src="./img/zadig_2.png" width="400px">

    （このスクショは置き換え後に撮ったものなので、ボタンの表記が "Reinstall Driver" となっていますが気にしないでください）

1. 中央下の選択ボックスで `libusb-win32 (vX.X.X.X)` を選択します
    - 最初は `WinUSB` が選択されているはずですが動作しないので注意してください

1. `Replace Driver` ボタンを押して完了するまで待ちます
    - デバイスマネージャーを起動して確認してみてください

    <img alt="インストール後のデバイスマネージャー" src="./img/devmgmt.png" width="400px">


#### 2a. ビルド済みの実行ファイルを使う

1. Windows用の実行ファイル (mtplvcap_windows_amd64.zip) を[ここ](https://github.com/puhitaku/mtplvcap/releases)からダウンロードします
1. ZIPをダブルクリックして展開します
1. `mtplvcap.exe` をダブルクリックして実行します
    - カメラのシャッターが開くのを確認してください


#### 2b. MSYS2で手でビルドする

1. MSYS2を[ここ](https://www.msys2.org/)からダウンロードしてインストールします
1. スタートメニューから "MSYS2 MSYS" を実行します
1. 依存するパッケージをインストールします

    - パッケージグループ `mingw-w64-x86_64-toolchain` とその他のパッケージを一度にインストールしようとするとエラーになる報告が上がっていて、これらは別々にインストールしたほうが良いかもしれません

	```sh
    pacman -Sy
    pacman -S mingw-w64-x86_64-toolchain
    pacman -S \
        mingw-w64-x86_64-libusb \
        mingw-w64-x86_64-go \
        mingw-w64-x86_64-pkg-config \
        git
    ```

1. PATHを追加します

    ```sh
    echo 'PATH=$PATH:/mingw64/bin:/mingw64/lib/go/bin' >> ~/.bashrc
    source ~/.bashrc
    ```

1. このリポジトリをクローンします

    ```sh
    git clone https://github.com/puhitaku/mtplvcap.git
    ```

1. `cd`してビルドして起動します

    ```sh
    cd mtplvcap
    CGO_CFLAGS='-Wno-deprecated-declarations' go build .
    ./mtplvcap.exe -debug server
    ```

    - カメラのシャッターが開くのを確認してください
    - `GOROOT=/mingw64/lib/go go build .` のようにGOROOTの指定が要るかもしれません

1. ゴール！
    - ビルドされたバイナリは移動したり再配布したりすることが可能です
    - `C:\msys64\mingw64\bin\libusb-1.0.dll` から `libusb-1.0.dll` を `mtplvcap.exe` と同じディレクトリにコピーしてくればどこでも起動できます


### macOS

#### 1. 依存パッケージをインストールする
1. [Homebrew](https://brew.sh/)をインストールします

1. libusbをインストールします

    ```sh
    brew install libusb
    ```


#### 2a. ビルド済みの実行ファイルを使う

1. macOS用の実行ファイル (mtplvcap_macos_amd64.zip) を[ここ](https://github.com/puhitaku/mtplvcap/releases)からダウンロードします
1. ZIPを展開して起動します

    ```sh
    unzip mtplvcap_macos_amd64.zip
    ./macos/mtplvcap
    ```

    - 初回はセキュリティのアラートが出て実行できないはずです

1. 検証されていないバイナリの実行を許可する

    - 「キャンセル」でダイアログを消します (ゴミ箱には入れないでください！)

    <img alt="警告ダイアログ" src="./img/macos_warning_jp.png" width="400px">

    - 「システム環境設定」 -> 「セキュリティとプライバシー」 と開き、「このまま許可」 をクリック

    <img alt="システム環境設定" src="./img/macos_warning2_jp.png" width="400px">

    - `mtplvcap` ももう一度起動し「開く」をクリック

    <img alt="警告ダイアログ2" src="./img/macos_warning3_jp.png" width="400px">

    - カメラのシャッターが開くのを確認してください


#### 2b. 手でビルドする

1. 依存パッケージをインストールします

    ```sh
    brew install golang git
    ```

1. XCode Command Line Toolsをインストールします

    ```sh
    xcode-select --install
    ```

1. このリポジトリをクローンします

    ```sh
    git clone https://github.com/puhitaku/mtplvcap.git
    ```

1. `cd`してビルドして起動します

    ```sh
    cd mtplvcap
    CGO_CFLAGS='-Wno-deprecated-declarations' go build .
    ./mtplvcap -debug server
    ```

    - カメラのシャッターが開くのを確認してください

1. ゴール！


### Linux

ビルド済みバイナリはありますが、Linuxはディストリによって環境が大きく異なるため、手でビルドすることをおすすめします。


#### 1. 依存パッケージをインストールする

1. libusbをインストールします

    ```sh
    # For Debian/Ubuntu:
    sudo apt install libusb-1.0.0
    ```

    ```sh
    # For Arch:
    pacman --sync libusb
    ```


#### 2a. ビルド済みの実行ファイルを使う

1. Linux用の実行ファイル (mtplvcap_linux_amd64.zip) を[ここ](https://github.com/puhitaku/mtplvcap/releases)からダウンロードします

1. ZIPを展開して起動します

    ```sh
    unzip mtplvcap_linux_amd64.zip
    ./linux/mtplvcap
    ```

    - カメラのシャッターが開くのを確認してください


#### 2b. 手でビルドする

1. 依存パッケージをインストールします

    ```sh
    # For Debian/Ubuntu:
    sudo apt install golang-go libusb-1.0.0-dev
    ```

    ```sh
    # For Arch:
    pacman --sync go libusb
    ```

1. `cd`してビルドして起動します

    ```sh
    cd mtplvcap
    CGO_CFLAGS='-Wno-deprecated-declarations' go build .
    ./mtplvcap -debug server
    ```

    - カメラのシャッターが開くのを確認してください


### 使い方

```sh
$ ./mtplvcap -help
Usage of ./mtplvcap:
  -backend-go
        force gousb as libusb wrapper (not recommended)
  -debug string
        comma-separated list of debugging options: usb, data, mtp, server
  -host string
        hostname: default = localhost, specify 0.0.0.0 for public access (default "localhost")
  -port int
        port: default = 42839 (default 42839)
  -product-id string
        PID of the camera to search (in hex), default=0x0 (all) (default "0x0")
  -server-only
        serve frontend without opening a DSLR (for devevelopment)
  -vendor-id string
        VID of the camera to search (in hex), default=0x0 (all) (default "0x0")
```


#### 撮られている映像を見る

 - `http://localhost:42839/view` を開くとキャプチャされたフレームが見えます


#### ブラウザでカメラを制御する

 - `http://localhost:42839` を開くとカメラを制御するコントローラーが使えます
 - "Auto Focus" セクションは一定間隔もしくは手動でAFを動作させられます
 - "Rate Limit" セクションはフレームレートの上限を設定でき、CPU消費量の削減に使えます
 - "Information" セクションはキャプチャされているフレームの大きさ、FPS、プレビューが見えます


#### Zoom, Google Meet, Google Hangoutsなどとつなぐ

1. mtplvcapをインストールし、動作することを確認します

1. OBS (Open Broadcaster Software) を[ここ](https://obsproject.com/)からインストールします

1. OBSバーチャルカメラをインストールします（OSにより異なります; ググってください）

1. OBSの設定を開き「映像」タブを開きます

1. Live Viewの画像サイズとぴったり合うように映像サイズを設定します
    - mtplvcapを起動して `localhost:42839` を開くと「Information」セクションにサイズが載っています

    <img alt="コントローラー" src="./img/obs_1.png" width="400px">
    <img alt="解像度設定" src="./img/obs_2.png" width="400px">

1. 「ブラウザ」ソースを追加します

    <img alt="ブラウザソースを追加" src="./img/obs_3.png" width="400px">

1. 「幅」と「高さ」の値を「映像」タブの設定と同じ値に設定します

1. URLを `http://localhost:42839/view` に設定します

    <img alt="URLを設定" src="./img/obs_4.png" width="400px">

1. バーチャルカメラをONにしてチャットアプリを設定します

1. キター！！！

    <img alt="Hi!" src="./img/obs_5.png" width="400px">
    <img alt="Zoom!" src="./img/obs_6.png" width="400px">


### 諸注意

 - このソフトはアルファ版です


### 既知の問題

 - 勝手にLVが止まる
    - この自動オフは「パワーオフ時間」の設定で延長できますが、ものによって最大30分だったり、無制限だったりします
        - D5300の場合: "カスタムメニュー" -> "c AEロック・タイマー" -> "c2 パワーオフ時間" -> "カスタマイズ" -> "ライブビュー表示" -> "30分"
    - 現状では、LVが止まっていたら自動で開始するというワークアラウンドで対処しています
 - Windowsのみ: MinTTY（MSYS2付属のターミナルエミュレータ）でmtplvcapを動かしている時、Ctrl-Cで終了すると終了処理が行われずに突然Killされる
    - 終了処理が行われないので、次の起動時に初期化に失敗したり、ケーブルの抜き差しが必要になったりします
    - これはMinTTYおよびでは知られた動作であり、mtplvcapのバグではありません
    - winptyをpacmanでインストールして、それ経由で起動すると解決します: `pacman -Sy winpty && winpty ./mtplvcap`
    - Explorerから直接mtplvcapを起動するのは問題ありません、ただしバツボタンではなく必ずCtrl-Cで終了してください


### フィードバック

 - IssueもPRも大歓迎です。[CONTRIBUTING.md](./CONTRIBUTING.md)に従って提出してください。
 - まだごく一部の機種しか動作確認できていません。是非お手持ちのカメラが動作したかどうか教えてください。何卒！


### Special Thanks

このプログラムは[github.com/hanwen/go-mtpfs](https://github.com/hanwen/go-mtpfs)からForkして大改造したものです。
go-mtpfsの成熟したMTP実装のおかげで実装のスタートを切れました。Han-Wenさん本当にありがとうございます。

[github.com/dukus/digiCamControl](https://github.com/dukus/digiCamControl)もMTPのペイロードをパースするために大いに参考にしました。
もしこのコードがなかったらmtplvcapは実装できていなかったと思います。

### ライセンス

[LICENSEファイルはこちら](./LICENSE)
