/*
Copyright 2024 KubeWorkz Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubeconfig

import "testing"

func TestLoadKubeConfigFromBytes(t *testing.T) {
	var KubeConfig = `
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM1ekNDQWMrZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJeE1ETXhPREE0TlRFeE9Gb1hEVE14TURNeE5qQTROVEV4T0Zvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBS0JWCkUwTHJ1eUlhNzQ1YmJkK1pIZDdTSEFnSmNPK3pKRWQ2QjBRakVGR3JOcnpjWGthM0ZkVWhhUXlFcEhCZ0FoZFAKQlVGRDVibms2REFhVHBGa1RTY2h6WTdISDZYRkpuTXhRWTlvdm1BQVlpeEFla0xZS2FQTVlGU3lUc1FtcUZjbwpnMFR2VVF1djBYNU4wandTYVh3bXJyZlRwMWZ2eTNNRzhzaHhXVW9YaVRvMEdVMTNIOGRPY2wxelVpSjJUWDNjCmtTVVFZZ1BBN3gzZ1llTEp3RnpTVlNsVFUrQXlLT1hKNFcxenAyVnVKOHVmTkNZYU5rdE54UWpxTi9vaFM0YjAKNjNQWmF1dXBPbkJBLzZLU2hYbUlCZThMTzZaQmg5d2dZSW85RzRCbjRmek5peGpmSWdpUS8xTEdwbStEU1RQUgpRZzczWnZVODMwcjQ0Nk0vSU84Q0F3RUFBYU5DTUVBd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZBZkkwcnBLdWhlZDNDM3d6bS9QaGlTZUgzVURNQTBHQ1NxR1NJYjMKRFFFQkN3VUFBNElCQVFBTS93dVgxWGxLOEVEbUx2RWlwdmFPK0FiNmxndlptcW5KL1FDWGZKRlJQeEZTeDlFVgpqSWNtQytEU1dvenl2c2t2ZWhoTVVVN2REQjhMVWx5aitGUEJkL25nUVJBZjFkUzVPTldncFo1UlZqTXhzTi9yCk9WZmwwWDZnME9hbUFoMHZmYzlRT3BYWXU5aUxpSVVvV1BPbjdTUWRVd1gwM25yWGl2Skx2SGM1dFRGWHN1dUsKNkxiOHJMcG1tYm1IL1ZlNW95ZmRKNk1xTnhkYklQQXM5dzdMWXZEODM3Z1dmUUVrL3VYaU1IcmoydCs1RzJsegpwU1hPZ3JmMUtmRG50R1JlWHYzVmZKcHYrcTdPcE1jTWx4NkZEZkZLekVwdmp2TDR3ZnNBV3l2MXlSd3F1ZVAxClNpYVg4TTFtQWlkK2FLdWtHZVBpN25KQVNPZU9aajJkSWFPNgotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    server: https://10.173.32.133:6443
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: kubernetes-admin
  name: kubernetes-admin@kubernetes
current-context: kubernetes-admin@kubernetes
kind: Config
preferences: {}
users:
- name: kubernetes-admin
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURFekNDQWZ1Z0F3SUJBZ0lJRnArTEcvWmcvdm93RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB5TVRBek1UZ3dPRFV4TVRoYUZ3MHlNakF6TVRnd09EVXhNakZhTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQW9GeUFVVVhBNElVMXFnUHMKUVdDVnNha1k0Tkg4cXVpaDZ1RGZPelRTT3h6UHhia2g4dnd2V1J5TzBTTndidm42d01YUFJtelVvbGN6THltTwpEcCtFWjlNeGpQQjZuT0Nvblg1V2RWNzByajljWnNjak5ENkZ0UjlIS3RkQVVoSEFoSWZMM1VhVURhbGFEVzBHCjM1cy8yYkFWbm1Wam1kVXQvUzArenVTSnZ4UmVWaW9QV2pPR1kydTViVGxSVk5VK05ocm9Oc1hCdEhxMzZpSjYKUjJGVjVhYXVWbFZidTRIL0tqQ1BNNEZuNlJmSWFzUzBJS2JlSnJadm5uWXBrOHFadFRVd0V0YVlmSUdDd2JrZwpxWEJjazFMTnlydjJoaGhNYWZ2Um1nOHhTekx5WS9QVWtXL1JiblRYcmZCQTZpQ2dEcGt2SGVOSmMvdFh5YmcwCjBKLzFPUUlEQVFBQm8wZ3dSakFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0h3WURWUjBqQkJnd0ZvQVVCOGpTdWtxNkY1M2NMZkRPYjgrR0pKNGZkUU13RFFZSktvWklodmNOQVFFTApCUUFEZ2dFQkFHVENWUklTckdnVmJIYU9wSzlnZXdlK1ZwaDd0Mm5ZWlJHbExjN0FLUU5GQmVJSW5MbWZodVhXCklQZUF3RDhBUlBCNmFHSE8ySHcrSVh5T0ZPZWp6bExlczBPZ29UQlc5ZU1mSHNLd25XNVlUYXBSbGRNSCtJdzMKK0xLRzdFby8wZmY3NDV0cklrbGw0UkI1cjkyNEMrbWtUQXNsekxNUElta0tKTTEzczMwZUtGOHpMbFhOOTl6egpBUEk2aDJaL2paaUZXMGp1MVFHL25iUGUzQXBvOUs1SDljdzFZemw5Q1lvYUpia3RveTViSzYxbU5sYkcveWltCjJsR1E4c0ZKL01HaGZxU0NzK0ZEK21TWHg5SVBWMUN0SW1oOGtpZzQwSXFKKzgvSXpnNHlJdGwwVlRjaC9GYnkKRUIvRWx5c2xUWjNjeFJyRTg3MXBaL2lUTmNaR2dTUT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb2dJQkFBS0NBUUVBb0Z5QVVVWEE0SVUxcWdQc1FXQ1ZzYWtZNE5IOHF1aWg2dURmT3pUU094elB4YmtoCjh2d3ZXUnlPMFNOd2J2bjZ3TVhQUm16VW9sY3pMeW1PRHArRVo5TXhqUEI2bk9Db25YNVdkVjcwcmo5Y1pzY2oKTkQ2RnRSOUhLdGRBVWhIQWhJZkwzVWFVRGFsYURXMEczNXMvMmJBVm5tVmptZFV0L1MwK3p1U0p2eFJlVmlvUApXak9HWTJ1NWJUbFJWTlUrTmhyb05zWEJ0SHEzNmlKNlIyRlY1YWF1VmxWYnU0SC9LakNQTTRGbjZSZklhc1MwCklLYmVKclp2bm5ZcGs4cVp0VFV3RXRhWWZJR0N3YmtncVhCY2sxTE55cnYyaGhoTWFmdlJtZzh4U3pMeVkvUFUKa1cvUmJuVFhyZkJBNmlDZ0Rwa3ZIZU5KYy90WHliZzAwSi8xT1FJREFRQUJBb0lCQUhIdnh5djNoNGIrbnBaaApteWNJWE5PUjlaOG5FNExMTHVBWFRnUmZEMC80dEpjalpyK2g5bkkySERMMEh4cDZlbk1sR0pSTkZ2Y1JSY2lvCm1jcENCRzFRWE5CcXZITmlHK3RxckR0UWNFQ3R2QlU2UUFVS3R5MXRQNzlzbU1LMjRqWkgxYzB1TEZ0WWpDY2wKNDlCVUdoV3RTbTcwVXNRbDl6ci9QclQ1SS9XaWYwaDR1UklIenF2Qkk2Vk1Id2NKS05acTByME5lWnJmZWNiZApzK2l2U1k3elkxbk1CNW5jdC83cElDcW9PS1EwcEhnWnFwZElmTlJGY1hrMytIT2pXYmhrWklHYWN5Y3lka0dMCktxUGJZNXVrVHozaTBRSFUzZXcrNmtPSjhQWms1VVRPaml5eDBTV3IybkJWWXBtaDJjQkFUZHFKUnFsdEU3MXQKeGJnV0dWRUNnWUVBeTRWb1BZOCtleUpLWFZRVitKWGdHS1Y5bU9DN2hrdEtEZXREVVV4ME1kVzNyQjZjSit2TgppK2dpd1NUamM4UjVraXVzUTVKcC9ZdTNscW9GUFRaUnNSMjEyU2RmMTFKSmxWdUl2T0tsUkhhRGs4UktDalJNCk1HNEV0dUZCUExaV2U3dU1raWRLQjNibE5Eb1I4U1lKUFFHSUxXSy8xdDg4SWRYVXRNK2ZQdThDZ1lFQXliWVUKNm5sTWhYQ3RLVTZablI4TzY5RjM4TUNoWFAxVzBMV1Q2YTBnTVROZzJEclpmL1RqNjBuSjVxcWU1elc0V3hQawpsMjVQeEdpVVB4RkpJY2l6NGJBZHNZamlCZzN0eVhHSFY1T25EYkdyS3N1UUtzc0dVa3ZtRjJOQjZaV0N2YmxoCnJlUFRQdVd5M2lzT2J5ekVaVGQwU1I3Mkk1aDhvd21WOFF6M2psY0NnWUFYT2I4K256VTdLaHVnY1BNYzdrV24KcW1NZDZpK2NVTVUvdDJSMFI2eE83NXZKV2FqcWFWK0lvdElzaG9BcWV2YUF0dGt1ME91bGtxYzkyMk5EckFwQQpreXBvZ0xUUVJzUzg5Ymd5RGc5Y004WXFPOUZUUHNxZkVEOGJlN05OTVdYSE13MEV0TlVlNnZPWE5UVk05dEZCCkZBYXRYMEtUUytCNTRBUXBLalI3SXdLQmdCRElpckw3eHNjdm1lamU5bDhoYjI5bHJYSWx4UWRrdU8yQ3ZqenEKUDh4NE51Mm03K1A2cjJBcmNwWXp5aWI0ekU4ZnQ1eDEvRU1lWkg5ZTQ5UGd6RDdTRlpacENrMXdxVUZpcGQwKwpsdDdIMHJNcjN1SUFwSlVDWmJHNmU0aUEvVERtRk4rUUhrcVVlQzZPTEpSSmlFblh0R0JGS0R2TCswcmhpRTFYClE5M2ZBb0dBTlRwSWhxaW9SUmRpa0JJQkcrYlptK21jZEFYU0EveFZZQmhSd2kwcHlLY1lTbU56L0w0RExDMjUKVUZjMVA1Y2ZqYnFXR01vWFJ3WXBkQUIyVVNKQmNFYmN0Q21jaUlkM2FhdzdYT2kwTWNaTkpraFlkZUpyUEFVNgpzRFVLa3p1WjlQRGFKc0s4c2VFVnBUeUhzeS84N0dHZ1NpSUtDRVlqN2U3RXNtSDRQd1U9Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==
`

	config, err := LoadKubeConfigFromBytes([]byte(KubeConfig))
	if err != nil {
		t.Fatal(err)
	}

	if config.CAData == nil {
		t.Fatal("no ca data")
	}
}
