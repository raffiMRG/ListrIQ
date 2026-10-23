# Rencana Anggaran Biaya (RAB) Perangkat Monitoring Listrik

Dokumen ini adalah estimasi biaya perangkat untuk satu titik ukur listrik rumah. Harga dalam Rupiah merupakan perkiraan harga eceran dan perlu dikonfirmasi kembali kepada penjual sebelum pembelian.

## Perangkat yang Digunakan

- PZEM-004T v3 beserta current transformer (CT) untuk membaca tegangan, arus, daya, energi, frekuensi, dan faktor daya.
- ESP32 DevKitC V4 dengan ESP32-WROOM-32U sebagai mikrokontroler.
- NRF24L01+ PA/LNA sebagai komunikasi radio antara Sensor Node dan Gateway Node.

## RAB Arsitektur Penuh

Arsitektur ini mengikuti rancangan Sensor Node (ESP32 + PZEM + NRF24L01) dan Gateway Node (ESP32 + NRF24L01), sehingga diperlukan dua ESP32 dan dua modul radio.

| No. | Komponen                  | Spesifikasi                                         |     Qty | Harga Satuan |      Subtotal |
| --- | ------------------------- | --------------------------------------------------- | ------: | -----------: | ------------: |
| 1   | + CT                      | Modul meter listrik AC dengan CT split-core         |       1 |     Rp75.000 |      Rp75.000 |
| 2   | ESP32 DevKitC V4          | ESP32-WROOM-32U, antena eksternal/U.FL              |       2 |     Rp65.000 |     Rp130.000 |
| 3   | NRF24L01+ PA/LNA          | Modul 2,4 GHz dengan antena eksternal               |       2 |     Rp25.000 |      Rp50.000 |
| 4   | Adaptor NRF24L01          | Soket adaptor dengan regulator 3,3 V                |       2 |      Rp6.000 |      Rp12.000 |
| 5   | Kapasitor elektrolit      | 10-100 uF untuk stabilisasi catu NRF24L01           |       2 |      Rp2.000 |       Rp4.000 |
| 6   | Adaptor catu daya 5 V     | Minimal 2 A, untuk Sensor Node dan Gateway Node     |       2 |     Rp45.000 |      Rp90.000 |
| 7   | Box/enclosure panel       | Box plastik ABS untuk melindungi rangkaian          |       2 |     Rp70.000 |     Rp140.000 |
| 8   | Kabel jumper dan kabel DC | Pengkabelan internal perangkat                      | 1 paket |     Rp40.000 |      Rp40.000 |
| 9   | Terminal block/konektor   | Sambungan kabel catu daya dan sensor                | 1 paket |     Rp25.000 |      Rp25.000 |
| 10  | Bahan pendukung           | Heat-shrink, cable tie, sekrup, dan isolasi listrik | 1 paket |     Rp30.000 |      Rp30.000 |
|     | **Total Estimasi**        |                                                     |         |              | **Rp596.000** |

## Alternatif RAB MVP WiFi Langsung

Implementasi firmware saat ini juga menyediakan mode MVP: satu ESP32 membaca PZEM-004T dan mengirim data langsung ke MQTT melalui WiFi. Pada mode ini, NRF24L01 dan Gateway Node tidak diperlukan.

| No. | Komponen                    | Spesifikasi                                 | Harga Estimasi | URL Sumber                                                       | Gambar                                                                                                                                                                                                                                                                                                                                    |
| --- | --------------------------- | ------------------------------------------- | -------------: | ---------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | PZEM-004T v4 100A+ CT       | Modul meter listrik AC dengan CT split-core |      Rp180.000 | [PZEM-004T v4](https://tk.tokopedia.com/ZSgmhYrKf/)              | ![PZEM-004T + CT](https://p16-images-sign-sg.tokopedia-static.net/tos-alisg-i-aphluv4xwc-sg/a28d18478a8c4954b00b8b751d3e4173~tplv-aphluv4xwc-white-pad-v1:1600:1600.jpeg?lk3s=0ccea506&x-expires=1792779026&x-signature=NlH5db0w8T5WwLKw2GBEu2pMo3A%3D&x-signature-webp=PxhVt%2BFKhGBfLugRpUysO3zJqMk%3D)                                 |
| 2   | ESP32 DevKitC V4            | ESP32-WROOM-32U                             |       Rp90.000 | [ESP32-DevKitC V4](https://tk.tokopedia.com/ZSgmkFwng/)          | ![ESP32 DevKitC V4](https://p16-images-sign-sg.tokopedia-static.net/tos-alisg-i-aphluv4xwc-sg/img/hDjmkQ/2023/10/10/e9a0c7cd-6d70-4f8c-bc7d-ac4a82e08b34.jpg~tplv-aphluv4xwc-white-pad-v1:1600:1600.jpeg?lk3s=0ccea506&x-expires=1792779456&x-signature=2tb%2FohJWAjI1nLe6ARmEFo1icNo%3D&x-signature-webp=UmOOpGhR3xNN6ALGTjufFvBICVw%3D) |
| 3   | Adaptor catu daya 5 V       | Minimal 2 A                                 |       Rp40.000 | [Referensi adaptor 5 V 2 A](https://tk.tokopedia.com/ZSgmkeqUA/) | ![ROBOT Charger RT-L1 ](https://p16-images-sign-sg.tokopedia-static.net/tos-alisg-i-aphluv4xwc-sg/331559d44495440fbab1dff7d3c00618~tplv-aphluv4xwc-white-pad-v1:1600:1600.jpeg?lk3s=0ccea506&x-expires=1792779565&x-signature=4YHxmnnXw5zUPvcQQ7YbFt3YQdg%3D&x-signature-webp=aNW9dW05GE%2BUTSpPG5ujy1KXnM0%3D)                           |
| 6   | Broco Steker Arde 13310     | Broco Steker Arde 13310                     |       Rp12.000 | [Referensi terminal block](https://tk.tokopedia.com/ZSgmkfBGJ/)  | ![Broco Steker Arde 13310 ](https://p16-images-sign-sg.tokopedia-static.net/tos-alisg-i-aphluv4xwc-sg/362c094ab4ed47678e25e022ed64a7a2~tplv-aphluv4xwc-white-pad-v1:1600:1600.jpeg?lk3s=0ccea506&x-expires=1792779819&x-signature=69tBqA9cooxGQo4fW7YiyY27W98%3D&x-signature-webp=hLx4q1EweI4U5KIPfvAS2kr%2BYWo%3D)                       |
|     | **Total Estimasi MVP WiFi** |                                             |  **Rp322.000** |                                                                  |                                                                                                                                                                                                                                                                                                                                           |

## Catatan

- Harga belum termasuk ongkos kirim, jasa pemasangan, komputer/server, router WiFi, dan biaya operasional listrik.
- Pastikan varian PZEM-004T yang dibeli kompatibel dengan CT bawaan dan kapasitas arus instalasi yang diukur.
- Gunakan enclosure tertutup, kabel berisolasi baik, serta terminal yang sesuai untuk instalasi panel listrik.
- PZEM-004T terhubung ke tegangan PLN. Matikan MCB sebelum pemasangan dan gunakan teknisi listrik yang kompeten bila diperlukan.
