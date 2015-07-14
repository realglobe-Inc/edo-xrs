# EDO xRS について

 EDO xRS is a store of experience record using [Experience API (aka, xAPI)](http://experienceapi.com/).   
If you dont know what is xAPI, see [xAPI-Spec](https://github.com/adlnet/xAPI-Spec/) for detail.  

 EDO xRS (Experience Recored Store) は[米国ADL](http://www.adlnet.org/)が策定している 
[Experience API (xAPI)](http://experienceapi.com/) を  
使ったスケーラブルで高速なパーソナルレコード保管システムです。  
本プロジェクトはパーソナルレコードフレームワーク [EDO](https://github.com/realglobe-Inc/edo) の一部として開発されています。

 EDO については [EDOについて](https://github.com/realglobe-Inc/edo/wiki) をご参照ください。  
[xAPI](https://github.com/elc-gh/xAPI-Spec_ja) ついては 
[e-Learming Consortium Japan](http://www.elc.or.jp/) がその[仕様の日本語版を公開](https://github.com/elc-gh/xAPI-Spec_ja)しています。

# 背景
 EDO xRS (Experience Record Store) は EDO において目指しているユーザーデータの  
シームレスな相互運用の目的の一部として, 個人から発生する様々なデータを  
記録するためのプラットフォームとして開発されました。  
EDO xRS は学習記録や健康/医療分野等, 膨大なデータを扱う必要があるため,  
ハイパフォーマンスなアプリケーションを目指して開発されています。

# 利用方法

### xAPI 対応バージョン

- 本システムは xAPI, Version 1.0.2 に対応しています。

### エンドポイント例

- http://edoxrs-server.example.com/{ユーザー名}/{アプリケーション名}/statements

**※ {ユーザー名}, {アプリケーション名} は適切な値に置き換えてご利用ください**

注: 本システムは将来的には xAPI のフルサポートを目指していますが,  
現在は Statement API のみの実装に留まっています。
詳細は以下の Issue をご覧ください。

https://github.com/realglobe-Inc/edo-xrs/issues/1

### リクエストサンプル
 本サーバーへリクエストを発行する例を以下に示す。
* sample.json
```json
{
    "actor": {
        "objectType": "Agent",
        "name": "Taro Realglobe",
        "account": {
            "homePage": "http://realglobe.example.com",
            "name": "Taro Realglobe"
        }
    },
    "verb": {
        "id": "http://www.adlnet.gov/XAPIprofile/ran(travelled_a_distance)",
        "display": {
            "ja-JP": "走った",
            "en-US": "ran"
        }
    },
    "object": {
        "objectType": "StatementRef",
        "id": "1cabcb4f-c41c-49a5-ad89-9a9c8c5fd20a"
    }
}
```

* ステートメントの挿入例
```sh
$ curl -X POST -d '@sample.json' -H 'X-Experience-API-Version: 1.0.2' http://127.0.0.1:3000/test/test/statements
["184f2820-1e2a-48f9-ade2-e7e5ad731fea"]
```

* 挿入したステートメントの取得例
```sh
$ curl -X GET -H 'X-Experience-API-Version: 1.0.2' http://127.0.0.1:3000/test/test/statements?statementId=184f2820-1e2a-48f9-ade2-e7e5ad731fea

{"actor":{"account":{"homePage":"http://realglobe.example.com","name":"Taro Realglobe"},"name":"Taro Realglobe","objectType":"Agent"},"id":"184f2820-1e2a-48f9-ade2-e7e5ad731fea","object":{"id":"1cabcb4f-c41c-49a5-ad89-9a9c8c5fd20a","objectType":"StatementRef"},"stored":"2015-07-06T08:33:04.464Z","verb":{"display":{"en-US":"ran","ja-JP":"走った"},"id":"http://www.adlnet.gov/XAPIprofile/ran(travelled_a_distance)"}}
```

### 実行方法
 本サーバーの構築方法は以下の通りです。

#### 必要条件
* [go (1.4.2)](https://golang.org)
* [PCRE (8.3.7)](http://www.pcre.org)
* [mongodb (3.0.4)](http://mongodb.org)

#### Vagrant
* 仮想環境構築ツール, Vagrant を使った EDO LRS サーバーの立ち上げ方
* vagrant が入っている環境において, 以下のコマンドを入力

```sh
$ vagrant up
$ # 正常に終了すると, http://192.168.33.10:3000 にて xRS サーバーが待機しています
```

#### 手動構築
* go, mongodb が実行可能な環境において, 以下のコマンドを入力

```sh
$ mkdir edo-xrs; cd edo-xrs
$ export GOPATH=$PWD
$ go get github.com/realglobe-Inc/edo-xrs
$
$ ./bin/edo-xrs
$ # 正常に起動すると, http://127.0.0.1:3000 にて xRS サーバーが待機しています
```

# ライセンス
   Copyright &copy;2015 Realglobe, Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

# お問い合わせ
本ソフトウェアに関して、ご意見、ご感想がございましたら株式会社リアルグローブまでお問い合わせください。  
Please contuct us if you are interested in this product.

http://realglobe.jp/

<a href="http://realglobe.jp">
  <img src="http://realglobe.jp/img/rg-logo.png" width="300px" alt="Realglobe, Inc."/>
</a>
